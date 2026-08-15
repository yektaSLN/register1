package service

import (
	"errors"

	"login/models"
	"login/repository"
	"login/utils"

	"gorm.io/gorm"
)

type ProductService interface {
	Create(userID uint, product *models.Product) error
	GetByID(userID uint, productID uint) (*models.Product, error)
	GetAll(userID uint) ([]models.Product, error)
	Update(userID uint, productID uint, product *models.Product) error
	Delete(userID uint, productID uint) error
}

type productService struct {
	productRepository repository.ProductRepository
}

func NewProductService(
	productRepository repository.ProductRepository,
) ProductService {
	return &productService{
		productRepository: productRepository,
	}
}

func (s *productService) Create(
	userID uint,
	product *models.Product,
) error {

	if product == nil {
		return utils.ErrInvalidRequest
	}

	// Never trust user_id coming from the client.
	// The authenticated user's ID comes from JWT.
	product.UserID = userID

	return s.productRepository.Create(product)
}

func (s *productService) GetByID(
	userID uint,
	productID uint,
) (*models.Product, error) {

	product, err := s.productRepository.FindByID(
		productID,
		userID,
	)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrProductNotFound
		}

		return nil, err
	}

	return product, nil
}

func (s *productService) GetAll(
	userID uint,
) ([]models.Product, error) {

	return s.productRepository.FindAll(userID)
}

func (s *productService) Update(
	userID uint,
	productID uint,
	product *models.Product,
) error {

	if product == nil {
		return utils.ErrInvalidRequest
	}

	existingProduct, err := s.productRepository.FindByID(
		productID,
		userID,
	)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrProductNotFound
		}

		return err
	}

	// Do not allow changing ownership.
	existingProduct.Name = product.Name
	existingProduct.Price = product.Price

	return s.productRepository.Update(existingProduct)
}

func (s *productService) Delete(
	userID uint,
	productID uint,
) error {

	// First check ownership/existence.
	_, err := s.productRepository.FindByID(
		productID,
		userID,
	)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrProductNotFound
		}

		return err
	}

	return s.productRepository.Delete(
		productID,
		userID,
	)
}
