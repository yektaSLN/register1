package service

import (
	"context"
	"errors"

	"login/kafka"
	"login/models"
	"login/repository"
	"login/utils"

	"github.com/rs/zerolog/log"
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
	kafkaProducer     *kafka.Producer
}

func NewProductService(
	productRepository repository.ProductRepository,
	kafkaProducer *kafka.Producer,
) ProductService {
	return &productService{
		productRepository: productRepository,
		kafkaProducer:     kafkaProducer,
	}
}

func (s *productService) Create(
	userID uint,
	product *models.Product,
) error {

	if product == nil {
		return utils.ErrInvalidRequest
	}

	product.UserID = userID

	if err := s.productRepository.Create(product); err != nil {
		return err
	}

	eventPayload := map[string]any{
		"id":      product.ID,
		"user_id": product.UserID,
		"name":    product.Name,
		"price":   product.Price,
	}

	if err := s.kafkaProducer.Publish(
		context.Background(),
		kafka.EventProductCreated,
		eventPayload,
	); err != nil {
		log.Error().
			Err(err).
			Uint("product_id", product.ID).
			Uint("user_id", userID).
			Str("event", kafka.EventProductCreated).
			Msg("failed to publish kafka event")
	}

	return nil
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

	existingProduct.Name = product.Name
	existingProduct.Price = product.Price

	if err := s.productRepository.Update(existingProduct); err != nil {
		return err
	}

	eventPayload := map[string]any{
		"id":      existingProduct.ID,
		"user_id": existingProduct.UserID,
		"name":    existingProduct.Name,
		"price":   existingProduct.Price,
	}

	if err := s.kafkaProducer.Publish(
		context.Background(),
		kafka.EventProductUpdated,
		eventPayload,
	); err != nil {
		log.Error().
			Err(err).
			Uint("product_id", existingProduct.ID).
			Uint("user_id", userID).
			Str("event", kafka.EventProductUpdated).
			Msg("failed to publish kafka event")
	}

	return nil
}

func (s *productService) Delete(
	userID uint,
	productID uint,
) error {

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

	if err := s.productRepository.Delete(
		productID,
		userID,
	); err != nil {
		return err
	}

	eventPayload := map[string]any{
		"id":      productID,
		"user_id": userID,
	}

	if err := s.kafkaProducer.Publish(
		context.Background(),
		kafka.EventProductDeleted,
		eventPayload,
	); err != nil {
		log.Error().
			Err(err).
			Uint("product_id", productID).
			Uint("user_id", userID).
			Str("event", kafka.EventProductDeleted).
			Msg("failed to publish kafka event")
	}

	return nil
}
