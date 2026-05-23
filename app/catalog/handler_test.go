package catalog

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type mockProductRepository struct {
	products       []models.Product
	total          int64
	err            error
	capturedFilter models.ProductFilter
}

func (m *mockProductRepository) GetProducts(filter models.ProductFilter) ([]models.Product, int64, error) {
	m.capturedFilter = filter
	return m.products, m.total, m.err
}

func (m *mockProductRepository) GetProductByCode(code string) (*models.Product, error) {
	return nil, nil
}

func TestHandleGet(t *testing.T) {
	t.Run("returns 200 with products list including category", func(t *testing.T) {
		mock := &mockProductRepository{
			products: []models.Product{
				{
					Code:     "PROD001",
					Price:    decimal.NewFromFloat(10.99),
					Category: models.Category{Code: "clothing", Name: "Clothing"},
				},
			},
			total: 1,
		}
		handler := NewCatalogHandler(mock)
		req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		rec := httptest.NewRecorder()
		handler.HandleGet(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.JSONEq(t,
			`{"products":[{"code":"PROD001","price":10.99,"category":{"code":"clothing","name":"Clothing"}}],"total":1,"offset":0,"limit":10}`,
			rec.Body.String(),
		)
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		mock := &mockProductRepository{err: assert.AnError}
		handler := NewCatalogHandler(mock)
		req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		rec := httptest.NewRecorder()
		handler.HandleGet(rec, req)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.JSONEq(t, `{"error":"internal server error"}`, rec.Body.String())
	})

	t.Run("passes offset and limit to repository", func(t *testing.T) {
		mock := &mockProductRepository{total: 0}
		handler := NewCatalogHandler(mock)
		req := httptest.NewRequest(http.MethodGet, "/catalog?offset=5&limit=20", nil)
		rec := httptest.NewRecorder()
		handler.HandleGet(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, 5, mock.capturedFilter.Offset)
		assert.Equal(t, 20, mock.capturedFilter.Limit)
	})

	t.Run("returns 400 when limit is out of range", func(t *testing.T) {
		mock := &mockProductRepository{}
		handler := NewCatalogHandler(mock)
		req := httptest.NewRequest(http.MethodGet, "/catalog?limit=200", nil)
		rec := httptest.NewRecorder()
		handler.HandleGet(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"error":"limit must be between 1 and 100"}`, rec.Body.String())
	})

	t.Run("returns 400 when offset is negative", func(t *testing.T) {
		mock := &mockProductRepository{}
		handler := NewCatalogHandler(mock)
		req := httptest.NewRequest(http.MethodGet, "/catalog?offset=-1", nil)
		rec := httptest.NewRecorder()
		handler.HandleGet(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"error":"offset must be a non-negative integer"}`, rec.Body.String())
	})

	t.Run("returns 400 when max_price is not a number", func(t *testing.T) {
		mock := &mockProductRepository{}
		handler := NewCatalogHandler(mock)
		req := httptest.NewRequest(http.MethodGet, "/catalog?max_price=abc", nil)
		rec := httptest.NewRecorder()
		handler.HandleGet(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"error":"max_price must be a number"}`, rec.Body.String())
	})

	t.Run("passes category filter to repository", func(t *testing.T) {
		mock := &mockProductRepository{total: 0}
		handler := NewCatalogHandler(mock)
		req := httptest.NewRequest(http.MethodGet, "/catalog?category=clothing", nil)
		rec := httptest.NewRecorder()
		handler.HandleGet(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "clothing", mock.capturedFilter.Category)
	})

	t.Run("passes max_price filter to repository", func(t *testing.T) {
		mock := &mockProductRepository{total: 0}
		handler := NewCatalogHandler(mock)
		req := httptest.NewRequest(http.MethodGet, "/catalog?max_price=50.0", nil)
		rec := httptest.NewRecorder()
		handler.HandleGet(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotNil(t, mock.capturedFilter.MaxPrice)
		assert.Equal(t, 50.0, *mock.capturedFilter.MaxPrice)
	})
}
