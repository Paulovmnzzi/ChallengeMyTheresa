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
	product        *models.Product
	productErr     error
}

func (m *mockProductRepository) GetProducts(filter models.ProductFilter) ([]models.Product, int64, error) {
	m.capturedFilter = filter
	return m.products, m.total, m.err
}

func (m *mockProductRepository) GetProductByCode(code string) (*models.Product, error) {
	return m.product, m.productErr
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

func TestHandleGetByCode(t *testing.T) {
	newMux := func(handler *CatalogHandler) *http.ServeMux {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /catalog/{code}", handler.HandleGetByCode)
		return mux
	}

	t.Run("returns 200 with product, category and variants", func(t *testing.T) {
		mock := &mockProductRepository{
			product: &models.Product{
				Code:     "PROD001",
				Price:    decimal.NewFromFloat(10.99),
				Category: models.Category{Code: "clothing", Name: "Clothing"},
				Variants: []models.Variant{
					{Name: "Variant A", SKU: "SKU001A", Price: decimal.NewFromFloat(11.99)},
					{Name: "Variant B", SKU: "SKU001B", Price: decimal.NewFromFloat(10.99)},
				},
			},
		}
		handler := NewCatalogHandler(mock)
		req := httptest.NewRequest(http.MethodGet, "/catalog/PROD001", nil)
		rec := httptest.NewRecorder()
		newMux(handler).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.JSONEq(t,
			`{"code":"PROD001","price":10.99,"category":{"code":"clothing","name":"Clothing"},"variants":[{"name":"Variant A","sku":"SKU001A","price":11.99},{"name":"Variant B","sku":"SKU001B","price":10.99}]}`,
			rec.Body.String(),
		)
	})

	t.Run("returns 404 when product not found", func(t *testing.T) {
		mock := &mockProductRepository{product: nil}
		handler := NewCatalogHandler(mock)
		req := httptest.NewRequest(http.MethodGet, "/catalog/NOTEXIST", nil)
		rec := httptest.NewRecorder()
		newMux(handler).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.JSONEq(t, `{"error":"product not found"}`, rec.Body.String())
	})

	t.Run("variant without price inherits product price", func(t *testing.T) {
		mock := &mockProductRepository{
			product: &models.Product{
				Code:  "PROD001",
				Price: decimal.NewFromFloat(10.99),
				Variants: []models.Variant{
					{Name: "V1", SKU: "SKU1", Price: decimal.NewFromFloat(0)},
				},
			},
		}
		handler := NewCatalogHandler(mock)
		req := httptest.NewRequest(http.MethodGet, "/catalog/PROD001", nil)
		rec := httptest.NewRecorder()
		newMux(handler).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t,
			`{"code":"PROD001","price":10.99,"category":{"code":"","name":""},"variants":[{"name":"V1","sku":"SKU1","price":10.99}]}`,
			rec.Body.String(),
		)
	})

	t.Run("variant with own price does not inherit product price", func(t *testing.T) {
		mock := &mockProductRepository{
			product: &models.Product{
				Code:  "PROD001",
				Price: decimal.NewFromFloat(10.99),
				Variants: []models.Variant{
					{Name: "V1", SKU: "SKU1", Price: decimal.NewFromFloat(25.00)},
				},
			},
		}
		handler := NewCatalogHandler(mock)
		req := httptest.NewRequest(http.MethodGet, "/catalog/PROD001", nil)
		rec := httptest.NewRecorder()
		newMux(handler).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t,
			`{"code":"PROD001","price":10.99,"category":{"code":"","name":""},"variants":[{"name":"V1","sku":"SKU1","price":25}]}`,
			rec.Body.String(),
		)
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		mock := &mockProductRepository{productErr: assert.AnError}
		handler := NewCatalogHandler(mock)
		req := httptest.NewRequest(http.MethodGet, "/catalog/PROD001", nil)
		rec := httptest.NewRecorder()
		newMux(handler).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.JSONEq(t, `{"error":"internal server error"}`, rec.Body.String())
	})
}
