package models

type ProductRepository interface {
	GetProducts(filter ProductFilter) ([]Product, int64, error)
	GetProductByCode(code string) (*Product, error)
}

type ProductFilter struct {
	Category      string
	PriceLessThan *float64
	Offset        int
	Limit         int
}
