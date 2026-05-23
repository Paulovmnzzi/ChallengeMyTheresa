package categories

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type CategoriesHandler struct {
	repo models.CategoryRepository
}

func NewCategoriesHandler(r models.CategoryRepository) *CategoriesHandler {
	return &CategoriesHandler{repo: r}
}

type categoryResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type listResponse struct {
	Categories []categoryResponse `json:"categories"`
}

type createRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (h *CategoriesHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	cats, err := h.repo.GetAllCategories()
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := listResponse{Categories: make([]categoryResponse, len(cats))}
	for i, c := range cats {
		resp.Categories[i] = categoryResponse{Code: c.Code, Name: c.Name}
	}

	api.OKResponse(w, resp)
}

func (h *CategoriesHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		api.ErrorResponse(w, http.StatusUnsupportedMediaType, "content-type must be application/json")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" || name == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "code and name are required")
		return
	}

	cat := &models.Category{Code: code, Name: name}
	if err := h.repo.CreateCategory(cat); err != nil {
		if errors.Is(err, models.ErrDuplicateCode) {
			api.ErrorResponse(w, http.StatusConflict, "category code already exists")
			return
		}
		api.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(categoryResponse{Code: cat.Code, Name: cat.Name})
}
