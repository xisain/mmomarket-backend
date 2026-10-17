package application

import (
	"context"
	"errors"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"xisain/mmomarket-backend/internal/platform/database/db"
)

type testRow struct {
	value any
	err   error
}

func (r testRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	value := reflect.ValueOf(r.value)
	for i := range dest {
		reflect.ValueOf(dest[i]).Elem().Set(value.Field(i))
	}
	return nil
}

type queryStep struct {
	contains string
	row      testRow
}
type testDB struct {
	db.DBTX
	t        *testing.T
	steps    []queryStep
	lastArgs []any
}

func (d *testDB) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	d.lastArgs = args
	d.t.Helper()
	if len(d.steps) == 0 {
		d.t.Fatal("unexpected database query")
		return testRow{}
	}
	step := d.steps[0]
	d.steps = d.steps[1:]
	if !strings.Contains(query, step.contains) {
		d.t.Fatalf("query %q does not contain %q", query, step.contains)
	}
	return step.row
}
func testUUID(n byte) pgtype.UUID { return pgtype.UUID{Bytes: [16]byte{n}, Valid: true} }
func validListingInput() RegisterListingInput {
	return RegisterListingInput{UserID: testUUID(1), SellerID: testUUID(2), GameID: testUUID(3), CategoryID: testUUID(4),
		Name: "Gold", Slug: "game-gold", Price: pgtype.Numeric{Int: big.NewInt(1000), Exp: -2, Valid: true},
		StockType: "finite", StockQuantity: pgtype.Int8{Int64: 5, Valid: true}, QuantityPerUnit: 1}
}
func TestRegisterListingValidation(t *testing.T) {
	tests := []struct {
		name   string
		change func(*RegisterListingInput)
		want   error
	}{
		{"trimmed name", func(i *RegisterListingInput) { i.Name = "  a  " }, ErrListingInvalidName},
		{"slug", func(i *RegisterListingInput) { i.Slug = "bad slug" }, ErrListingInvalidSlug},
		{"missing game", func(i *RegisterListingInput) { i.GameID = pgtype.UUID{} }, ErrListingInvalidReference},
		{"negative price", func(i *RegisterListingInput) { i.Price.Int = big.NewInt(-1) }, ErrListingInvalidPrice},
		{"nonfinite price", func(i *RegisterListingInput) { i.Price.NaN = true }, ErrListingInvalidPrice},
		{"missing finite stock", func(i *RegisterListingInput) { i.StockQuantity.Valid = false }, ErrListingInvalidStock},
		{"negative stock", func(i *RegisterListingInput) { i.StockQuantity.Int64 = -1 }, ErrListingInvalidStock},
		{"unlimited with quantity", func(i *RegisterListingInput) { i.StockType = "unlimited" }, ErrListingInvalidStock},
		{"quantity per unit", func(i *RegisterListingInput) { i.QuantityPerUnit = 0 }, ErrListingInvalidStock},
		{"server without region", func(i *RegisterListingInput) { i.ServerID = testUUID(5) }, ErrListingInvalidRegion},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validListingInput()
			tt.change(&input)
			service := &ListingService{queries: db.New(&testDB{t: t})}
			_, err := service.RegisterListing(context.Background(), input)
			if !errors.Is(err, tt.want) {
				t.Fatalf("got %v, want %v", err, tt.want)
			}
		})
	}
}
func TestSellerAuthorization(t *testing.T) {
	for _, tc := range []struct {
		name     string
		owner    pgtype.UUID
		status   string
		queryErr error
		want     error
	}{
		{"approved owner", testUUID(1), "approved", nil, nil},
		{"other owner", testUUID(9), "approved", nil, ErrListingSellerNotAllowed},
		{"pending owner", testUUID(1), "pending", nil, ErrListingSellerNotAllowed},
		{"rejected owner", testUUID(1), "rejected", nil, ErrListingSellerNotAllowed},
		{"missing seller", pgtype.UUID{}, "", pgx.ErrNoRows, ErrListingSellerNotAllowed},
		{"database failure", pgtype.UUID{}, "", errors.New("database unavailable"), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			database := &testDB{t: t, steps: []queryStep{{"FROM sellers", testRow{value: db.Seller{OwnerID: tc.owner, Status: tc.status}, err: tc.queryErr}}}}
			service := &ListingService{queries: db.New(database)}
			err := service.authorizeSeller(context.Background(), testUUID(1), testUUID(2))
			if tc.queryErr != nil && tc.want == nil {
				if !errors.Is(err, tc.queryErr) {
					t.Fatalf("lost database error: %v", err)
				}
				return
			}
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}
func TestRegisterListing(t *testing.T) {
	for _, stockType := range []string{"finite", "unlimited"} {
		t.Run(stockType, func(t *testing.T) {
			input := validListingInput()
			input.StockType = stockType
			if stockType == "unlimited" {
				input.StockQuantity = pgtype.Int8{}
			}
			database := &testDB{t: t, steps: []queryStep{
				{"FROM sellers", testRow{value: db.Seller{OwnerID: input.UserID, Status: "approved"}}},
				{"INSERT INTO listings", testRow{value: db.Listing{ID: testUUID(10), SellerID: input.SellerID, Status: "draft", StockType: stockType}}},
			}}
			service := &ListingService{queries: db.New(database)}
			listing, err := service.RegisterListing(context.Background(), input)
			if len(database.lastArgs) != 14 || database.lastArgs[4] != input.SellerID || database.lastArgs[13] != "draft" || database.lastArgs[10] != stockType || database.lastArgs[11] != input.StockQuantity {
				t.Fatalf("incorrect insert arguments: %v", database.lastArgs)
			}
			if err != nil || listing.Status != "draft" || listing.SellerID != input.SellerID || len(database.steps) != 0 {
				t.Fatalf("listing=%+v, err=%v", listing, err)
			}
		})
	}
}
func TestLocationRelationships(t *testing.T) {
	for _, tc := range []struct {
		name  string
		steps []queryStep
		want  error
	}{
		{"wrong game", []queryStep{{"FROM game_regions", testRow{value: db.GameRegion{GameID: testUUID(9)}}}}, ErrListingInvalidRegion},
		{"missing region", []queryStep{{"FROM game_regions", testRow{err: pgx.ErrNoRows}}}, ErrListingInvalidRegion},
		{"wrong server region", []queryStep{{"FROM game_regions", testRow{value: db.GameRegion{GameID: testUUID(3)}}}, {"FROM game_servers", testRow{value: db.GameServer{RegionID: testUUID(9)}}}}, ErrListingInvalidServer},
		{"matching", []queryStep{{"FROM game_regions", testRow{value: db.GameRegion{GameID: testUUID(3)}}}, {"FROM game_servers", testRow{value: db.GameServer{RegionID: testUUID(5)}}}}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &ListingService{queries: db.New(&testDB{t: t, steps: tc.steps})}
			err := service.validateLocation(context.Background(), testUUID(3), testUUID(5), testUUID(6))
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}
func TestRegisterCategory(t *testing.T) {
	for _, tc := range []struct {
		name, slug     string
		queryErr, want error
	}{
		{"  a  ", "gold", nil, ErrListingCategoryInvalidName},
		{strings.Repeat("a", 101), "gold", nil, ErrListingCategoryInvalidName},
		{"Gold", "bad slug", nil, ErrListingCategoryInvalidSlug},
		{"Gold", "gold", &pgconn.PgError{Code: "23505", ConstraintName: "listing_categorys_slug_key"}, ErrListingCategorySlugTaken},
		{"Gold", "gold", nil, nil},
	} {
		t.Run(tc.name+tc.slug, func(t *testing.T) {
			database := &testDB{t: t, steps: []queryStep{{"INSERT INTO listing_categorys", testRow{value: db.ListingCategory{Name: "Gold", Slug: "gold"}, err: tc.queryErr}}}}
			service := &ListingService{queries: db.New(database)}
			_, err := service.RegisterListingCategory(context.Background(), RegisterListingCategoryInput{Name: tc.name, Slug: tc.slug})
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

func TestSellerListingsRejectOtherOwner(t *testing.T) {
	for _, operation := range []string{"create", "list"} {
		t.Run(operation, func(t *testing.T) {
			input := validListingInput()
			database := &testDB{t: t, steps: []queryStep{{"FROM sellers", testRow{value: db.Seller{OwnerID: testUUID(9), Status: "approved"}}}}}
			service := &ListingService{queries: db.New(database)}
			var err error
			if operation == "create" {
				_, err = service.RegisterListing(context.Background(), input)
			} else {
				_, err = service.ListSellerListings(context.Background(), input.UserID, db.ListListingsBySellerParams{SellerID: input.SellerID, Limit: 20})
			}
			if !errors.Is(err, ErrListingSellerNotAllowed) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestListingDatabaseErrors(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want error
	}{
		{pgx.ErrNoRows, ErrListingNotFound},
		{&pgconn.PgError{Code: "23505", ConstraintName: "listings_slug_key"}, ErrListingAlreadyExists},
		{&pgconn.PgError{Code: "23503"}, ErrListingInvalidReference},
		{&pgconn.PgError{Code: "23514", ConstraintName: "valid_listing_stock"}, ErrListingInvalidStock},
	} {
		if got := listingError("create", tc.err); !errors.Is(got, tc.want) {
			t.Fatalf("got %v, want %v", got, tc.want)
		}
	}
}
