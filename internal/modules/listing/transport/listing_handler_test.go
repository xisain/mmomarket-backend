package transport

import (
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"xisain/mmomarket-backend/internal/modules/listing/application"
	"xisain/mmomarket-backend/internal/platform/database/db"
)

func TestParsePrice(t *testing.T) {
	for _, tc := range []struct {
		value string
		cents int64
	}{{"0", 0}, {"12.3", 1230}, {" 12.34 ", 1234}, {"9999999999.99", 999999999999}} {
		price, err := parsePrice(tc.value)
		if err != nil || price.Exp != -2 || price.Int.Cmp(big.NewInt(tc.cents)) != 0 {
			t.Fatalf("%q: price=%+v err=%v", tc.value, price, err)
		}
	}
	for _, value := range []string{"", "-1", "NaN", "Infinity", "1.234", "10000000000", "1e3"} {
		if _, err := parsePrice(value); err == nil {
			t.Fatalf("accepted invalid price %q", value)
		}
	}
}
func TestListingRoutesRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterListingRoutes(router.Group("/api/v1"), NewListingHandler(nil), nil)
	for _, route := range []struct{ method, path string }{
		{"POST", "/api/v1/admin/listing-categories"}, {"GET", "/api/v1/admin/listing-categories"},
		{"POST", "/api/v1/seller/123/listings"}, {"GET", "/api/v1/seller/123/listings"},
		{"GET", "/api/v1/seller/123/listing-categories"},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(route.method, route.path, nil))
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s: got %d", route.method, route.path, response.Code)
		}
	}
}
func TestRegisterListingReadsAuthUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("authUser", db.User{ID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true}}) })
	router.POST("/seller/:seller_id/listings", NewListingHandler(nil).RegisterListing)
	response := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/seller/00000000-0000-0000-0000-000000000002/listings", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("got %d: %s", response.Code, response.Body.String())
	}
}
func TestListingErrorResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		err    error
		status int
	}{
		{application.ErrListingSellerNotAllowed, 403}, {application.ErrListingCategorySlugTaken, 409},
		{application.ErrListingAlreadyExists, 409}, {application.ErrListingInvalidServer, 400},
		{application.ErrListingInvalidStock, 400}, {application.ErrListingInvalidPagination, 400},
		{application.ErrListingNotFound, 404}, {errors.New("secret database details"), 500},
	} {
		response := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(response)
		respondError(ctx, tc.err)
		if response.Code != tc.status {
			t.Fatalf("%v: got %d", tc.err, response.Code)
		}
		if tc.status == 500 && strings.Contains(response.Body.String(), tc.err.Error()) {
			t.Fatal("internal error exposed")
		}
	}
}
