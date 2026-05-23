package models

import (
	"errors"

	"gorm.io/gorm"
)

type ProductsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{db: db}
}

func (r *ProductsRepository) GetProducts(filter ProductFilter) ([]Product, int64, error) {
	var products []Product
	var total int64

	q := r.db.Model(&Product{})

	if filter.Category != "" {
		q = q.Where("category_id = (SELECT id FROM categories WHERE code = ?)", filter.Category)
	}
	if filter.MaxPrice != nil {
		q = q.Where("products.price < ?", *filter.MaxPrice)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := q.Preload("Category").Order("products.id").Offset(filter.Offset).Limit(filter.Limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductsRepository) GetProductByCode(code string) (*Product, error) {
	var product Product
	result := r.db.Preload("Category").Preload("Variants").Where("code = ?", code).First(&product)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &product, nil
}
