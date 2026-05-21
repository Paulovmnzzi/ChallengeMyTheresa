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
	products []models.Product
	total    int64
	err      error
}

func (m *mockProductRepository) GetProducts(filter models.ProductFilter) ([]models.Product, int64, error) {
	return m.products, m.total, m.err
}

func (m *mockProductRepository) GetProductByCode(code string) (*models.Product, error) {
	return nil, nil
}

func TestHandleGet(t *testing.T) {
	t.Run("returns 200 with products list", func(t *testing.T) {
		mock := &mockProductRepository{
			products: []models.Product{
				{Code: "PROD001", Price: decimal.NewFromFloat(10.99)},
				{Code: "PROD002", Price: decimal.NewFromFloat(20.99)},
			},
			total: 2,
		}

		handler := NewCatalogHandler(mock)
		req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.JSONEq(t,
			`{"products":[{"code":"PROD001","price":10.99},{"code":"PROD002","price":20.99}]}`,
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
}
