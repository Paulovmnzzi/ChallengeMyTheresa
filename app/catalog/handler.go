package catalog

import (
	"net/http"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type categoryResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type product struct {
	Code     string           `json:"code"`
	Price    float64          `json:"price"`
	Category categoryResponse `json:"category"`
}

type Response struct {
	Products []product `json:"products"`
	Total    int64     `json:"total"`
	Offset   int       `json:"offset"`
	Limit    int       `json:"limit"`
}

type CatalogHandler struct {
	repo models.ProductRepository
}

func NewCatalogHandler(r models.ProductRepository) *CatalogHandler {
	return &CatalogHandler{repo: r}
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	filter, ok := parseFilter(w, r)
	if !ok {
		return
	}

	res, total, err := h.repo.GetProducts(filter)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	products := make([]product, len(res))
	for i, p := range res {
		products[i] = product{
			Code:  p.Code,
			Price: p.Price.InexactFloat64(),
			Category: categoryResponse{
				Code: p.Category.Code,
				Name: p.Category.Name,
			},
		}
	}

	api.OKResponse(w, Response{
		Products: products,
		Total:    total,
		Offset:   filter.Offset,
		Limit:    filter.Limit,
	})
}

type variantResponse struct {
	Name  string  `json:"name"`
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
}

type productDetailResponse struct {
	Code     string            `json:"code"`
	Price    float64           `json:"price"`
	Category categoryResponse  `json:"category"`
	Variants []variantResponse `json:"variants"`
}

func (h *CatalogHandler) HandleGetByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	p, err := h.repo.GetProductByCode(code)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if p == nil {
		api.ErrorResponse(w, http.StatusNotFound, "product not found")
		return
	}

	variants := make([]variantResponse, len(p.Variants))
	for i, v := range p.Variants {
		price := v.Price
		if price.IsZero() {
			price = p.Price
		}
		variants[i] = variantResponse{
			Name:  v.Name,
			SKU:   v.SKU,
			Price: price.InexactFloat64(),
		}
	}

	api.OKResponse(w, productDetailResponse{
		Code:  p.Code,
		Price: p.Price.InexactFloat64(),
		Category: categoryResponse{
			Code: p.Category.Code,
			Name: p.Category.Name,
		},
		Variants: variants,
	})
}

func parseFilter(w http.ResponseWriter, r *http.Request) (models.ProductFilter, bool) {
	q := r.URL.Query()
	filter := models.ProductFilter{
		Category: q.Get("category"),
		Offset:   0,
		Limit:    10,
	}

	if raw := q.Get("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			api.ErrorResponse(w, http.StatusBadRequest, "offset must be a non-negative integer")
			return filter, false
		}
		filter.Offset = v
	}

	if raw := q.Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 100 {
			api.ErrorResponse(w, http.StatusBadRequest, "limit must be between 1 and 100")
			return filter, false
		}
		filter.Limit = v
	}

	if raw := q.Get("max_price"); raw != "" {
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			api.ErrorResponse(w, http.StatusBadRequest, "max_price must be a number")
			return filter, false
		}
		filter.MaxPrice = &v
	}

	return filter, true
}
