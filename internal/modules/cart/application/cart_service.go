package application

import (
	"context"
	"errors"
	"xisain/mmomarket-backend/internal/platform/database/db"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrListingNotFound = errors.New("listing not found")
	ErrCartNotFound    = errors.New("cart not found")
)
type CartService struct {
	queries *db.Queries
}

func NewCartService(pool *pgxpool.Pool) *CartService {
	return &CartService{
		queries: db.New(pool),
	}
}


// func (c *CartService) registerCart(ctx context.Context) (db.Cart, error) { }
// 
func (c *CartService) GetCart(ctx context.Context, userID pgtype.UUID) (db.GetCartRow, error) {
	if !userID.Valid {
		return db.GetCartRow{}, ErrCartNotFound
	}
	cart, err := c.queries.GetCart(ctx, userID)
	if err != nil {
		return db.GetCartRow{}, err
	}
	return cart, nil
}