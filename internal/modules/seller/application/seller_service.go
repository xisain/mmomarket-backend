package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	db "xisain/mmomarket-backend/internal/platform/database/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSellerExists = errors.New("Seller Already Exists")
	ErrSlugTaken = errors.New("seller slug already taken")
	ErrSellerNotFound = errors.New("seller not found or invalid status")
	ErrInvalidSlug = errors.New("invalid slug")
	
)
var slugPattern = regexp.MustCompile(
	`^[a-z0-9]+(?:-[a-z0-9]+)*$`,
)
type SellerService struct {
	pool *pgxpool.Pool
	queries db.Queries
}
func NewSellerService(pool *pgxpool.Pool) *SellerService {
	return &SellerService{
		pool: pool,
		queries: *db.New(pool),
	}
}
type RegisterSellerInput struct {
	ShopName string
	Slug	 string
	Description string
}

func (s *SellerService) Register (ctx context.Context, userID pgtype.UUID, input RegisterSellerInput) (db.Seller,error) {
	input.ShopName = strings.TrimSpace(input.ShopName)
	input.Slug = strings.TrimSpace(input.Slug)
	input.Description = strings.TrimSpace(input.Description)
	
	if len(input.Slug) > 180 ||
	!slugPattern.MatchString(input.Slug){
		return db.Seller{}, ErrInvalidSlug
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.Seller{}, err
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	seller, err := qtx.CreateSeller(
		ctx,
		db.CreateSellerParams{
			OwnerID:     userID,
			ShopName:    input.ShopName,
			Slug:        input.Slug,
			Description: pgtype.Text{
				String: input.Description,
				Valid: true,
			},	
		},
	)
	if err != nil {
	var pgErr  *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
			case "seller_owner_id_unique" :
				return db.Seller{}, ErrSellerExists
			case "seller_slug_key" :
				return db.Seller{}, ErrSlugTaken
		}
	}
	return db.Seller{}, fmt.Errorf(
		"create seller %w", err,
		)
	}
	_, err = qtx.CreateSellerMember(
			ctx,
			db.CreateSellerMemberParams{
				SellerID: seller.ID,
				UserID:   userID,
			},
		)
		if err != nil {
			return db.Seller{}, fmt.Errorf(
				"create owner membership: %w", err,
			)
		}
	
		if err := tx.Commit(ctx); err != nil {
			return db.Seller{}, err
		}
	
		return seller, nil
	
}
func (s *SellerService) GetByOwner(ctx context.Context, userID pgtype.UUID)(db.Seller, error){
	seller, err := s.queries.GetSellerByOwnerID(
		ctx,
		userID,
	)
	if errors.Is(err, pgx.ErrNoRows){
		return db.Seller{}, ErrSellerNotFound
	}
	return seller, nil
}
func (s *SellerService) List(
	ctx context.Context,
) ([]db.Seller, error) {
	return s.queries.ListSellers(ctx)
}

func (s *SellerService) IsAdmin(
	ctx context.Context,
	userID pgtype.UUID,
) (bool, error) {
	return s.queries.IsGlobalAdmin(ctx, userID)
}

func (s *SellerService) Approve(
	ctx context.Context,
	id pgtype.UUID,
) (db.Seller, error) {
	seller, err := s.queries.ApproveSeller(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return db.Seller{}, ErrSellerNotFound
	}
	return seller, err
}

func (s *SellerService) Reject(
	ctx context.Context,
	id pgtype.UUID,
) (db.Seller, error) {
	seller, err := s.queries.RejectSeller(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return db.Seller{}, ErrSellerNotFound
	}
	return seller, err
}