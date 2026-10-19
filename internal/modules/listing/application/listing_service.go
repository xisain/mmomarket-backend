package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
	"xisain/mmomarket-backend/internal/platform/database/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrListingCategoryInvalidName   = errors.New("Nama Listing terlalu panjang atau kurang dari 3")
	ErrListingCategoryInvalidSlug   = errors.New("Slug Listing tidak valid")
	ErrListingCategorySlugTaken     = errors.New("Slug Listing sudah digunakan")
	ErrListingCategoryNotFound      = errors.New("Listing Category tidak ditemukan")
	ErrListingCategoryAlreadyExists = errors.New("Listing Category sudah ada")
	ErrListingSellerNotAllowed      = errors.New("hanya pemilik seller approved yang dapat mengakses listing")
	ErrListingNotFound              = errors.New("Listing tidak ditemukan")
	ErrListingInvalidPagination     = errors.New("limit harus antara 1 dan 100 dan offset tidak boleh negatif")
	ErrListingAlreadyExists         = errors.New("Listing sudah ada")
	ErrListingInvalidName           = errors.New("nama listing kurang dari 3 atau terlalu panjang")
	ErrListingInvalidSlug           = errors.New("slug listing tidak valid")
	ErrListingInvalidPrice          = errors.New("harga listing tidak valid")
	ErrListingInvalidStock          = errors.New("stok listing tidak valid")
	ErrListingInvalidRegion         = errors.New("region listing tidak valid")
	ErrListingInvalidServer         = errors.New("server listing tidak valid")
	ErrListingInvalidReference      = errors.New("game, category, atau seller tidak ditemukan")
)
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type ListingService struct {
	queries *db.Queries
}

func NewListingService(pool *pgxpool.Pool) *ListingService {
	return &ListingService{
		queries: db.New(pool),
	}
}

type RegisterListingCategoryInput struct {
	Name        string
	Slug        string
	Description string
}
type RegisterListingInput struct {
	UserID          pgtype.UUID
	GameID          pgtype.UUID
	CategoryID      pgtype.UUID
	GameRegionID    pgtype.UUID
	ServerID        pgtype.UUID
	SellerID        pgtype.UUID
	Name            string
	Slug            string
	Description     string
	ImageURL        pgtype.Text
	Price           pgtype.Numeric
	StockType       string
	StockQuantity   pgtype.Int8
	QuantityPerUnit int64
}

func (h *ListingService) RegisterListingCategory(ctx context.Context, input RegisterListingCategoryInput) (db.ListingCategory, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.TrimSpace(input.Slug)
	input.Description = strings.TrimSpace(input.Description)
	if utf8.RuneCountInString(input.Name) < 3 || utf8.RuneCountInString(input.Name) > 100 {
		return db.ListingCategory{}, ErrListingCategoryInvalidName
	}
	if len(input.Slug) < 3 || len(input.Slug) > 100 || !slugPattern.MatchString(input.Slug) {
		return db.ListingCategory{}, ErrListingCategoryInvalidSlug
	}
	lg, err := h.queries.CreateListingCategory(ctx,
		db.CreateListingCategoryParams{
			Name: input.Name,
			Slug: input.Slug,
			Description: pgtype.Text{
				String: input.Description,
				Valid:  true,
			},
		},
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "listing_categorys_slug_key" {
			return db.ListingCategory{}, ErrListingCategorySlugTaken
		}
		return db.ListingCategory{}, fmt.Errorf("failed to register listing category: %w", err)
	}

	return lg, nil
}

func (h *ListingService) RegisterListing(ctx context.Context, input RegisterListingInput) (db.Listing, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.TrimSpace(input.Slug)
	input.Description = strings.TrimSpace(input.Description)
	if utf8.RuneCountInString(input.Name) < 3 || utf8.RuneCountInString(input.Name) > 255 {
		return db.Listing{}, ErrListingInvalidName
	}
	if len(input.Slug) < 3 || len(input.Slug) > 255 || !slugPattern.MatchString(input.Slug) {
		return db.Listing{}, ErrListingInvalidSlug
	}
	if !input.GameID.Valid || !input.CategoryID.Valid || !input.SellerID.Valid {
		return db.Listing{}, ErrListingInvalidReference
	}
	if input.ServerID.Valid && !input.GameRegionID.Valid {
		return db.Listing{}, ErrListingInvalidRegion
	}
	if !input.Price.Valid ||
		input.Price.NaN ||
		input.Price.InfinityModifier != 0 ||
		input.Price.Int == nil ||
		input.Price.Int.Sign() < 0 {
		return db.Listing{}, ErrListingInvalidPrice
	}
	switch input.StockType {
	case "finite":
		if !input.StockQuantity.Valid ||
			input.StockQuantity.Int64 < 0 {
			return db.Listing{}, ErrListingInvalidStock
		}

	case "unlimited":
		if input.StockQuantity.Valid {
			return db.Listing{}, ErrListingInvalidStock
		}

	default:
		return db.Listing{}, ErrListingInvalidStock
	}
	if input.QuantityPerUnit <= 0 {
		return db.Listing{}, ErrListingInvalidStock
	}
	if err := h.authorizeSeller(ctx, input.UserID, input.SellerID); err != nil {
		return db.Listing{}, err
	}
	if err := h.validateLocation(ctx, input.GameID, input.GameRegionID, input.ServerID); err != nil {
		return db.Listing{}, err
	}
	listing, err := h.queries.CreateListing(ctx,
		db.CreateListingParams{
			GameID:       input.GameID,
			CategoryID:   input.CategoryID,
			GameRegionID: input.GameRegionID,
			ServerID:     input.ServerID,
			SellerID:     input.SellerID,
			Name:         input.Name,
			Slug:         input.Slug,
			Description: pgtype.Text{
				String: input.Description,
				Valid:  input.Description != "",
			},
			ImageUrl: pgtype.Text{
				String: input.ImageURL.String,
				Valid:  input.ImageURL.Valid,
			},
			Price:           input.Price,
			StockType:       input.StockType,
			StockQuantity:   input.StockQuantity,
			QuantityPerUnit: input.QuantityPerUnit,
			Status:          "draft",
		},
	)
	if err != nil {
		return db.Listing{}, listingError("register listing", err)
	}
	return listing, nil
}
func (h *ListingService) ListListingCategories(ctx context.Context) ([]db.ListingCategory, error) {
	return h.queries.ListListingCategories(ctx)
}

// authorizeSeller checks ownership and approval before exposing seller listings.
func (h *ListingService) authorizeSeller(ctx context.Context, userID, sellerID pgtype.UUID) error {
	if !userID.Valid || !sellerID.Valid {
		return ErrListingSellerNotAllowed
	}
	seller, err := h.queries.GetSellerByID(ctx, sellerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrListingSellerNotAllowed
	}
	if err != nil {
		return fmt.Errorf("get seller: %w", err)
	}
	if seller.OwnerID != userID || seller.Status != "approved" {
		return ErrListingSellerNotAllowed
	}
	return nil
}

func (h *ListingService) ListSellerListings(ctx context.Context, userID pgtype.UUID, input db.ListListingsBySellerParams) ([]db.Listing, error) {
	if err := validatePagination(input.Limit, input.Offset); err != nil {
		return nil, err
	}
	if err := h.authorizeSeller(ctx, userID, input.SellerID); err != nil {
		return nil, err
	}
	return h.queries.ListListingsBySeller(ctx, input)
}

func (h *ListingService) validateLocation(ctx context.Context, gameID, regionID, serverID pgtype.UUID) error {
	if regionID.Valid {
		region, err := h.queries.GetGameRegionByID(ctx, regionID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrListingInvalidRegion
		}
		if err != nil {
			return fmt.Errorf("get listing region: %w", err)
		}
		if region.GameID != gameID {
			return ErrListingInvalidRegion
		}
	}
	if serverID.Valid {
		if !regionID.Valid {
			return ErrListingInvalidRegion
		}
		server, err := h.queries.GetGameServerByID(ctx, serverID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrListingInvalidServer
		}
		if err != nil {
			return fmt.Errorf("get listing server: %w", err)
		}
		if server.RegionID != regionID {
			return ErrListingInvalidServer
		}
	}
	return nil
}

func (h *ListingService) GetListingByID(ctx context.Context, id pgtype.UUID) (db.Listing, error) {
	if !id.Valid {
		return db.Listing{}, ErrListingNotFound
	}
	listing, err := h.queries.GetListingByID(ctx, id)
	return listing, listingError("get listing by id", err)
}

func (h *ListingService) GetListingBySlug(ctx context.Context, slug string) (db.Listing, error) {
	slug = strings.TrimSpace(slug)
	if len(slug) < 2 || len(slug) > 255 || !slugPattern.MatchString(slug) {
		return db.Listing{}, ErrListingInvalidSlug
	}
	listing, err := h.queries.GetListingBySlug(ctx, slug)
	return listing, listingError("get listing by slug", err)
}

func (h *ListingService) ListListings(ctx context.Context, input db.ListListingsParams) ([]db.Listing, error) {
	if err := validatePagination(input.Limit, input.Offset); err != nil {
		return nil, err
	}
	return h.queries.ListListings(ctx, input)
}

func (h *ListingService) ListListingsByGameAndCategory(ctx context.Context, input db.ListListingsByGameAndCategoryParams) ([]db.Listing, error) {
	if !input.GameID.Valid || !input.CategoryID.Valid {
		return nil, ErrListingInvalidReference
	}
	if err := validatePagination(input.Limit, input.Offset); err != nil {
		return nil, err
	}
	return h.queries.ListListingsByGameAndCategory(ctx, input)
}

func (h *ListingService) ListListingsByGameAndRegion(ctx context.Context, input db.ListListingsByGameAndRegionParams) ([]db.Listing, error) {
	if !input.GameID.Valid {
		return nil, ErrListingInvalidReference
	}
	if !input.GameRegionID.Valid {
		return nil, ErrListingInvalidRegion
	}
	if err := validatePagination(input.Limit, input.Offset); err != nil {
		return nil, err
	}
	return h.queries.ListListingsByGameAndRegion(ctx, input)
}

func (h *ListingService) ListListingsByServer(ctx context.Context, input db.ListListingsByServerParams) ([]db.Listing, error) {
	if !input.ServerID.Valid {
		return nil, ErrListingInvalidServer
	}
	if err := validatePagination(input.Limit, input.Offset); err != nil {
		return nil, err
	}
	return h.queries.ListListingsByServer(ctx, input)
}

func (h *ListingService) ListListingsBySeller(ctx context.Context, input db.ListListingsBySellerParams) ([]db.Listing, error) {
	if !input.SellerID.Valid {
		return nil, ErrListingInvalidReference
	}
	if err := validatePagination(input.Limit, input.Offset); err != nil {
		return nil, err
	}
	return h.queries.ListListingsBySeller(ctx, input)
}

func (h *ListingService) ListListingByGameSlugAndCategorySlug(ctx context.Context,input db.ListListingsByGameSlugAndCategoryParams) ([]db.Listing, error){
	if input.Slug == "" || input.Slug_2  == ""{
		return nil, ErrListingInvalidSlug
	}
	if err := validatePagination(input.Limit, input.Offset); err != nil {
		return nil, err
	}
	return h.queries.ListListingsByGameSlugAndCategory(ctx, input)
}

// UpdateListingInput contains only the fields editable by UpdateListing.
type UpdateListingInput struct {
	GameID       pgtype.UUID
	CategoryID   pgtype.UUID
	GameRegionID pgtype.UUID
	ServerID     pgtype.UUID
	Name         string
	Slug         string
	Description  string
	ImageUrl     pgtype.Text
	Price        pgtype.Numeric
}

// UpdateListing replaces editable fields for a listing belonging to sellerID.
func (h *ListingService) UpdateListing(ctx context.Context, id, sellerID pgtype.UUID, input UpdateListingInput) (db.Listing, error) {
	if !id.Valid {
		return db.Listing{}, ErrListingNotFound
	}
	if !sellerID.Valid || !input.GameID.Valid || !input.CategoryID.Valid {
		return db.Listing{}, ErrListingInvalidReference
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.TrimSpace(input.Slug)
	input.Description = strings.TrimSpace(input.Description)
	if utf8.RuneCountInString(input.Name) < 3 || utf8.RuneCountInString(input.Name) > 255 {
		return db.Listing{}, ErrListingInvalidName
	}
	if len(input.Slug) < 3 || len(input.Slug) > 255 || !slugPattern.MatchString(input.Slug) {
		return db.Listing{}, ErrListingInvalidSlug
	}
	if input.ServerID.Valid && !input.GameRegionID.Valid {
		return db.Listing{}, ErrListingInvalidRegion
	}
	if !input.Price.Valid || input.Price.NaN || input.Price.InfinityModifier != 0 || input.Price.Int == nil || input.Price.Int.Sign() < 0 {
		return db.Listing{}, ErrListingInvalidPrice
	}
	if err := h.validateLocation(ctx, input.GameID, input.GameRegionID, input.ServerID); err != nil {
		return db.Listing{}, err
	}
	listing, err := h.queries.UpdateListing(ctx, db.UpdateListingParams{
		ID: id, SellerID: sellerID, GameID: input.GameID, CategoryID: input.CategoryID,
		GameRegionID: input.GameRegionID, ServerID: input.ServerID, Name: input.Name, Slug: input.Slug,
		Description: pgtype.Text{String: input.Description, Valid: input.Description != ""}, ImageUrl: input.ImageUrl, Price: input.Price,
	})
	return listing, listingError("update listing", err)
}

// DeleteListing deletes only a listing belonging to sellerID.
func (h *ListingService) DeleteListing(ctx context.Context, id, sellerID pgtype.UUID) error {
	if !id.Valid {
		return ErrListingNotFound
	}
	if !sellerID.Valid {
		return ErrListingInvalidReference
	}
	affected, err := h.queries.DeleteListing(ctx, db.DeleteListingParams{ID: id, SellerID: sellerID})
	if err != nil {
		return listingError("delete listing", err)
	}
	if affected == 0 {
		return ErrListingNotFound
	}
	return nil
}

func validatePagination(limit, offset int32) error {
	if limit <= 0 || offset < 0 {
		return ErrListingInvalidPagination
	}
	return nil
}

func listingError(operation string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrListingNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			if pgErr.ConstraintName == "listings_slug_key" {
				return ErrListingAlreadyExists
			}
		case "23503":
			return ErrListingInvalidReference
		case "23514":
			switch pgErr.ConstraintName {
			case "server_requires_region":
				return ErrListingInvalidRegion
			case "listings_price_check":
				return ErrListingInvalidPrice
			case "valid_listing_stock", "listings_stock_type_check", "listings_quantity_per_unit_check":
				return ErrListingInvalidStock
			}
		}
	}
	return fmt.Errorf("%s: %w", operation, err)
}

