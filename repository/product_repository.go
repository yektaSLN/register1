package repository

import (
	"login/models"

	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(product *models.Product) error
	FindByID(id uint, userID uint) (*models.Product, error)
	FindAll(userID uint) ([]models.Product, error)
	Update(product *models.Product) error
	Delete(id uint, userID uint) error
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{
		db: db,
	}
}

func (r *productRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) FindByID(id uint, userID uint) (*models.Product, error) {
	var product models.Product

	err := r.db.
		Where("id = ? AND user_id = ?", id, userID).
		First(&product).Error

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) FindAll(userID uint) ([]models.Product, error) {
	var products []models.Product

	err := r.db.
		Where("user_id = ?", userID).
		Find(&products).Error

	return products, err
}

func (r *productRepository) Update(product *models.Product) error {
	return r.db.Save(product).Error
}

func (r *productRepository) Delete(id uint, userID uint) error {
	return r.db.
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.Product{}).Error
}
