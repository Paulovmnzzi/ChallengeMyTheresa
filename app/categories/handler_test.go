package categories

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
)

type mockCategoryRepository struct {
	categories      []models.Category
	createErr       error
	getErr          error
	createdCategory *models.Category
}

func (m *mockCategoryRepository) GetAllCategories() ([]models.Category, error) {
	return m.categories, m.getErr
}

func (m *mockCategoryRepository) CreateCategory(c *models.Category) error {
	m.createdCategory = c
	return m.createErr
}

func TestHandleGet(t *testing.T) {
	t.Run("returns 200 with categories list", func(t *testing.T) {
		mock := &mockCategoryRepository{
			categories: []models.Category{
				{Code: "clothing", Name: "Clothing"},
				{Code: "shoes", Name: "Shoes"},
			},
		}
		handler := NewCategoriesHandler(mock)
		req := httptest.NewRequest(http.MethodGet, "/categories", nil)
		rec := httptest.NewRecorder()
		handler.HandleGet(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.JSONEq(t,
			`{"categories":[{"code":"clothing","name":"Clothing"},{"code":"shoes","name":"Shoes"}]}`,
			rec.Body.String(),
		)
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		mock := &mockCategoryRepository{getErr: assert.AnError}
		handler := NewCategoriesHandler(mock)
		req := httptest.NewRequest(http.MethodGet, "/categories", nil)
		rec := httptest.NewRecorder()
		handler.HandleGet(rec, req)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.JSONEq(t, `{"error":"internal server error"}`, rec.Body.String())
	})
}

func TestHandlePost(t *testing.T) {
	t.Run("returns 201 with created category", func(t *testing.T) {
		mock := &mockCategoryRepository{}
		handler := NewCategoriesHandler(mock)
		body := bytes.NewBufferString(`{"code":"bags","name":"Bags"}`)
		req := httptest.NewRequest(http.MethodPost, "/categories", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.HandlePost(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.JSONEq(t, `{"code":"bags","name":"Bags"}`, rec.Body.String())
	})

	t.Run("returns 400 when code is empty", func(t *testing.T) {
		mock := &mockCategoryRepository{}
		handler := NewCategoriesHandler(mock)
		body := bytes.NewBufferString(`{"code":"","name":"Bags"}`)
		req := httptest.NewRequest(http.MethodPost, "/categories", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.HandlePost(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"error":"code and name are required"}`, rec.Body.String())
	})

	t.Run("returns 400 when name is empty", func(t *testing.T) {
		mock := &mockCategoryRepository{}
		handler := NewCategoriesHandler(mock)
		body := bytes.NewBufferString(`{"code":"bags","name":""}`)
		req := httptest.NewRequest(http.MethodPost, "/categories", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.HandlePost(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"error":"code and name are required"}`, rec.Body.String())
	})

	t.Run("returns 409 when code already exists", func(t *testing.T) {
		mock := &mockCategoryRepository{createErr: models.ErrDuplicateCode}
		handler := NewCategoriesHandler(mock)
		body := bytes.NewBufferString(`{"code":"clothing","name":"Clothing"}`)
		req := httptest.NewRequest(http.MethodPost, "/categories", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.HandlePost(rec, req)
		assert.Equal(t, http.StatusConflict, rec.Code)
		assert.JSONEq(t, `{"error":"category code already exists"}`, rec.Body.String())
	})

	t.Run("trims whitespace from code and name before saving", func(t *testing.T) {
		mock := &mockCategoryRepository{}
		handler := NewCategoriesHandler(mock)
		body := bytes.NewBufferString(`{"code":"  bags  ","name":"  Bags  "}`)
		req := httptest.NewRequest(http.MethodPost, "/categories", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.HandlePost(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "bags", mock.createdCategory.Code)
		assert.Equal(t, "Bags", mock.createdCategory.Name)
	})

	t.Run("returns 415 when Content-Type is not application/json", func(t *testing.T) {
		mock := &mockCategoryRepository{}
		handler := NewCategoriesHandler(mock)
		body := bytes.NewBufferString(`{"code":"bags","name":"Bags"}`)
		req := httptest.NewRequest(http.MethodPost, "/categories", body)
		req.Header.Set("Content-Type", "text/plain")
		rec := httptest.NewRecorder()
		handler.HandlePost(rec, req)
		assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
		assert.JSONEq(t, `{"error":"content-type must be application/json"}`, rec.Body.String())
	})
}
