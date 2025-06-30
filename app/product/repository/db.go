package repository

import (
	"context"

	"github.com/PorcoGalliard/eCommerce-Microservice/models"
)

func (r *ProductRepository) FindProductByID(ctx context.Context, productID int64) (*models.Product, error) {
	var product models.Product
	err := r.Database.WithContext(ctx).Table("product").Where("id = ?", productID).Last(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) FindProductCategoryByID(ctx context.Context, productCategoryID int) (*models.ProductCategory, error) {
	var productCategory models.ProductCategory
	err := r.Database.WithContext(ctx).Table("product_category").Where("id = ?", productCategoryID).Last(&productCategory).Error
	if err != nil {
		return nil, err
	}
	return &productCategory, nil
}

func (r *ProductRepository) InsertProduct(ctx context.Context, product *models.Product) (int64, error) {
	err := r.Database.WithContext(ctx).Table("product").Create(product).Error
	if err != nil {
		return 0, err
	}
	return product.ID, nil
}

func (r *ProductRepository) InsertProductCategory(ctx context.Context, productCategory *models.ProductCategory) (int, error) {
	err := r.Database.WithContext(ctx).Table("product_category").Create(productCategory).Error
	if err != nil {
		return 0, err
	}
	return productCategory.ID, nil
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	err := r.Database.WithContext(ctx).Table("product").Save(product).Error
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *ProductRepository) UpdateProductCategory(ctx context.Context, productCategory *models.ProductCategory) (*models.ProductCategory, error) {
	err := r.Database.WithContext(ctx).Table("product_category").Save(productCategory).Error
	if err != nil {
		return nil, err
	}
	return productCategory, nil
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, productID int64) error {
	err := r.Database.WithContext(ctx).Table("product").Delete(&models.Product{}, productID).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *ProductRepository) DeleteProductCategory(ctx context.Context, productCategoryID int) error {
	err := r.Database.WithContext(ctx).Table("product_category").Delete(&models.ProductCategory{}, productCategoryID).Error
	if err != nil {
		return err
	}
	return nil
}