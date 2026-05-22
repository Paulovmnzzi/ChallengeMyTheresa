package models

import "errors"

var ErrDuplicateCode = errors.New("duplicate code")

type ProductRepository interface {
	GetProducts(filter ProductFilter) ([]Product, int64, error)
	GetProductByCode(code string) (*Product, error)
}

type CategoryRepository interface {
	GetAllCategories() ([]Category, error)
	CreateCategory(c *Category) error
}

type ProductFilter struct {
	Category      string
	PriceLessThan *float64
	Offset        int
	Limit         int
}
