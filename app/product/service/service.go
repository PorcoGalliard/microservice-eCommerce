package service

import (
	"context"

	"github.com/PorcoGalliard/eCommerce-Microservice/app/product/repository"
	"github.com/PorcoGalliard/eCommerce-Microservice/infrastructure/log"
	"github.com/PorcoGalliard/eCommerce-Microservice/models"
	"github.com/sirupsen/logrus"
)

type ProductService struct {
	ProductRepo repository.ProductRepository
}

func NewProductService(productRepo *repository.ProductRepository) *ProductService {
	return &ProductService{
		ProductRepo: *productRepo,
	}
}

func (s *ProductService) GetProductByID(ctx context.Context, productID int64) (*models.Product, error) {
	product, err := s.ProductRepo.GetProductByIDFromRedis(ctx, productID)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"ProductID": productID,
		}).Errorf("s.ProductRepo.GetProductByIDFromRedis got an error at %v", err)
	} else if product.ID != 0 {
		return product, nil
	}

	product, err = s.ProductRepo.FindProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	ctxConcurrent := context.WithValue(ctx, context.Background(), ctx.Value("request_id"))
	go func(ctx context.Context, product *models.Product, productID int64) {
		errConcurrent := s.ProductRepo.SetProductByID(ctx, product, productID)

		if errConcurrent != nil {
			log.Logger.Info("Kena di sini")
			log.Logger.WithFields(logrus.Fields{
				"product": product,
			}).Errorf("s.ProductRepo.SetProductByID got an error at %v", errConcurrent)
		}
	}(ctxConcurrent, product, productID)

	return product, nil
}

func (s *ProductService) GetProductCategoryByID(ctx context.Context, productCategoryID int) (*models.ProductCategory, error) {
	productCategory, err := s.ProductRepo.FindProductCategoryByID(ctx, productCategoryID)
	if err != nil {
		return nil, err
	}
	return productCategory, nil
}

func (s *ProductService) CreateNewProduct(ctx context.Context, product *models.Product) (int64, error) {
	productID, err := s.ProductRepo.InsertProduct(ctx, product)
	if err != nil {
		return 0, err
	}
	return productID, nil
}

func (s *ProductService) CreateNewProductCategory(ctx context.Context, productCategory *models.ProductCategory) (int, error) {
	productCategoryID, err := s.ProductRepo.InsertProductCategory(ctx, productCategory)
	if err != nil {
		return 0, err
	}
	return productCategoryID, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	updatedProduct, err := s.ProductRepo.UpdateProduct(ctx, product)
	if err != nil {
		return nil, err
	}
	return updatedProduct, nil
}

func (s *ProductService) UpdateProductCategory(ctx context.Context, productCategory *models.ProductCategory) (*models.ProductCategory, error) {
	updatedProductCategory, err := s.ProductRepo.UpdateProductCategory(ctx, productCategory)
	if err != nil {
		return nil, err
	}
	return updatedProductCategory, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, productID int64) error {
	if err := s.ProductRepo.DeleteProduct(ctx, productID); err != nil {
		return err
	}
	return nil
}

func (s *ProductService) DeleteProductCategory(ctx context.Context, productCategoryID int) error {
	if err := s.ProductRepo.DeleteProductCategory(ctx, productCategoryID); err != nil {
		return err
	}
	return nil
}