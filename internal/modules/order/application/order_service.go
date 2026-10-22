package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"math"
	"time"
	"xisain/mmomarket-backend/internal/platform/database/db"
)

var (
	ErrOrderNotFound                = errors.New("order not found")
	ErrOrderAlreadyExists           = errors.New("order already exists")
	ErrOrderListingNotFound         = errors.New("order listing not found")
	ErrOrderListingAlreadyExists    = errors.New("order listing already exists")
	ErrOrderListingNotBelongsToUser = errors.New("order listing does not belong to user")
	ErrOrderInvalidReference        = errors.New("invalid order buyer, seller, or listing")
	ErrOrderInvalidQuantity         = errors.New("order quantity must be positive")
	ErrOrderInvalidAmount           = errors.New("invalid order price or total amount")
	ErrOrderInvalidStatus           = errors.New("invalid order status")
)

type OrderService struct {
	queries *db.Queries
}

func NewOrderService(pool *pgxpool.Pool) *OrderService {
	return &OrderService{
		queries: db.New(pool),
	}
}

type RegisterOrderInput struct {
	BuyerID     pgtype.UUID
	SellerID    pgtype.UUID
	ListingID   pgtype.UUID
	Quantity    int32
	UnitPrice   int64
	TotalAmount int64
}

func (s *OrderService) RegisterOrder(ctx context.Context, input RegisterOrderInput) (db.Order, error) {
	if !input.BuyerID.Valid || !input.SellerID.Valid || !input.ListingID.Valid {
		return db.Order{}, ErrOrderInvalidReference
	}
	if input.Quantity <= 0 {
		return db.Order{}, ErrOrderInvalidQuantity
	}
	if input.UnitPrice < 0 || input.TotalAmount < 0 ||
		input.UnitPrice > math.MaxInt64/int64(input.Quantity) ||
		input.TotalAmount != input.UnitPrice*int64(input.Quantity) {
		return db.Order{}, ErrOrderInvalidAmount
	}
	order, err := s.queries.CreateOrder(ctx, db.CreateOrderParams{
		BuyerID: input.BuyerID, SellerID: input.SellerID, ListingID: input.ListingID,
		Quantity: input.Quantity, UnitPrice: input.UnitPrice, TotalAmount: input.TotalAmount, ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(24 * time.Hour),
			Valid: true,
		},
	})
	return order, orderError("register order", err)
}

func (s *OrderService) BuyerGetOrder(
    ctx context.Context,
    buyerID pgtype.UUID,
) ([]db.BuyerGetOrderRow, error) {
    if !buyerID.Valid {
        return nil, ErrOrderInvalidReference
    }

    orders, err := s.queries.BuyerGetOrder(ctx, buyerID)
    if err != nil {
        return nil, orderError("list buyer orders", err)
    }

    return orders, nil
}

func (s *OrderService) BuyerGetOrderByStatus(ctx context.Context, input db.BuyerGetOrderByStatusParams) ([]db.Order, error) {
	if !input.BuyerID.Valid {
		return nil, ErrOrderInvalidReference
	}
	if err := validateOrderStatus(input.Status); err != nil {
		return nil, err
	}
	orders, err := s.queries.BuyerGetOrderByStatus(ctx, input)
	return orders, orderError("list buyer orders by status", err)
}

func (s *OrderService) SellerGetOrder(ctx context.Context, sellerID pgtype.UUID) ([]db.SellerGetOrderRow, error) {
	if !sellerID.Valid {
		return nil, ErrOrderInvalidReference
	}
	orders, err := s.queries.SellerGetOrder(ctx, sellerID)
	return orders, orderError("list seller orders", err)
}

func (s *OrderService) SellerGetOrderByStatus(ctx context.Context, input db.SellerGetOrderByStatusParams) ([]db.Order, error) {
	if !input.SellerID.Valid {
		return nil, ErrOrderInvalidReference
	}
	if err := validateOrderStatus(input.Status); err != nil {
		return nil, err
	}
	orders, err := s.queries.SellerGetOrderByStatus(ctx, input)
	return orders, orderError("list seller orders by status", err)
}

func (s *OrderService) UserUpdateOrder(ctx context.Context, input db.UserUpdateOrderParams) (db.Order, error) {
	if !input.ID.Valid {
		return db.Order{}, ErrOrderNotFound
	}
	if !input.BuyerID.Valid {
		return db.Order{}, ErrOrderInvalidReference
	}
	if err := validateOrderStatus(input.Status); err != nil {
		return db.Order{}, err
	}
	order, err := s.queries.UserUpdateOrder(ctx, input)
	return order, orderError("update buyer order", err)
}

func (s *OrderService) SellerUpdateOrder(ctx context.Context, input db.SellerUpdateOrderParams) (db.Order, error) {
	if !input.ID.Valid {
		return db.Order{}, ErrOrderNotFound
	}
	if !input.SellerID.Valid {
		return db.Order{}, ErrOrderInvalidReference
	}
	if err := validateOrderStatus(input.Status); err != nil {
		return db.Order{}, err
	}
	order, err := s.queries.SellerUpdateOrder(ctx, input)
	return order, orderError("update seller order", err)
}

func (s *OrderService) BuyerUpdateOrder(ctx context.Context, input db.UserUpdateOrderParams) (db.Order, error) {
	if !input.ID.Valid {
		return db.Order{}, ErrOrderNotFound
	}
	if !input.BuyerID.Valid {
		return db.Order{}, ErrOrderInvalidReference
	}
	if err := validateOrderStatus(input.Status); err != nil {
		return db.Order{}, err
	}
	order, err := s.queries.UserUpdateOrder(ctx, input)
	return order, orderError("update buyer order", err)
}

func (s *OrderService) BuyerFindOrderByID(ctx context.Context, input db.BuyerGetOrderByIDParams) (db.Order, error){
	if !input.ID.Valid {
		return db.Order{}, ErrOrderNotFound
	}
	order, err := s.queries.BuyerGetOrderByID(ctx, input)
	return order, orderError("find order by id", err)
}

func validateOrderStatus(status string) error {
	switch status {
	case "pending_payment", "paid", "processing", "delivered", "completed", "cancelled", "disputed", "refunded":
		return nil
	default:
		return ErrOrderInvalidStatus
	}
}

func orderError(operation string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrOrderNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrOrderAlreadyExists
		case "23503":
			if pgErr.ConstraintName == "orders_listing_id_fkey" {
				return ErrOrderListingNotFound
			}
			return ErrOrderInvalidReference
		case "23514":
			switch pgErr.ConstraintName {
			case "orders_status_check":
				return ErrOrderInvalidStatus
			case "orders_quantity_check":
				return ErrOrderInvalidQuantity
			case "orders_unit_price_check", "orders_total_amount_check":
				return ErrOrderInvalidAmount
			}
		}
	}
	return fmt.Errorf("%s: %w", operation, err)
}
