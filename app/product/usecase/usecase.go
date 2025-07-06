package usecase

import (
	"context"

	"github.com/PorcoGalliard/eCommerce-Microservice/app/product/service"
	"github.com/PorcoGalliard/eCommerce-Microservice/infrastructure/log"
	"github.com/PorcoGalliard/eCommerce-Microservice/models"
	"github.com/sirupsen/logrus"
)

type ProductUsecase struct {
	ProductService service.ProductService
}

func NewProductUsecase(productService *service.ProductService) *ProductUsecase {
	return &ProductUsecase{
		ProductService: *productService,
	}
}

func (uc *ProductUsecase) GetProductByID(ctx context.Context, productID int64) (*models.Product, error) {
	product, err := uc.ProductService.GetProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (uc *ProductUsecase) GetProductCategoryByID(ctx context.Context, productCategoryID int) (*models.ProductCategory, error) {
	productCategory, err := uc.ProductService.GetProductCategoryByID(ctx, productCategoryID)
	if err != nil {
		return nil, err
	}
	return productCategory, nil
}

func (uc *ProductUsecase) CreateNewProduct(ctx context.Context, product *models.Product) (int64, error) {
	productID, err := uc.ProductService.CreateNewProduct(ctx, product)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"name": product.Name,
			"category": product.Category_ID,
		}).Errorf("uc.CreateNewProduct got an error at %v", err)
		return 0, err
	}
	return productID, nil
}

func (uc *ProductUsecase) CreateNewProductCategory(ctx context.Context, productCategory *models.ProductCategory) (int, error) {
	productCategoryID, err := uc.ProductService.CreateNewProductCategory(ctx, productCategory)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"name": productCategory.Name,
		}).Errorf("uc.ProductService.CreateNewProductCategory got an error at %v", err)
		return 0, err
	}
	return productCategoryID, nil
}

func (uc *ProductUsecase) UpdateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	updatedProduct, err := uc.ProductService.UpdateProduct(ctx, product)
	if err != nil {
		return nil, err
	}
	return updatedProduct, nil
}

func (uc *ProductUsecase) UpdateProductCategory(ctx context.Context, productCategory *models.ProductCategory) (*models.ProductCategory, error) {
	updatedProductCategory, err := uc.ProductService.UpdateProductCategory(ctx, productCategory)
	if err != nil {
		return nil, err
	}
	return updatedProductCategory, nil
}

func (uc *ProductUsecase) DeleteProduct(ctx context.Context, productID int64) error {
	if err := uc.ProductService.DeleteProduct(ctx, productID); err != nil {
		return err
	}
	return nil
}

func (uc *ProductUsecase) DeleteProductCategory(ctx context.Context, productCategoryID int) error {
	if err := uc.ProductService.DeleteProductCategory(ctx, productCategoryID); err != nil {
		return err
	}
	return nil
}