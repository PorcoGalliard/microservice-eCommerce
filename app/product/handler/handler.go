package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/PorcoGalliard/eCommerce-Microservice/app/product/usecase"
	"github.com/PorcoGalliard/eCommerce-Microservice/infrastructure/log"
	"github.com/PorcoGalliard/eCommerce-Microservice/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ProductHandler struct {
	ProductUsecase usecase.ProductUsecase
}

func NewProductHandler(productUsecase *usecase.ProductUsecase) *ProductHandler {
	return &ProductHandler{
		ProductUsecase: *productUsecase,
	}
}

func (h *ProductHandler) ProductManagement(c *gin.Context) {
	var param *models.ProductManagementParameter
	if err := c.ShouldBindJSON(&param); err != nil {
		log.Logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "Invalid input",
		})
		return
	}

	if param.Action == "" {
		log.Logger.Error("❌ Missing Action")
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "Missing parameter action",
		})
		return
	}

	switch param.Action {
	case "add":
		if param.ID != 0 {
			log.Logger.Error("❌ Invalid request, params ID must be empty")
			c.JSON(http.StatusBadRequest, gin.H{
				"error_message": "Invalid request",
			})
			return
		}

		productID, err := h.ProductUsecase.CreateNewProduct(c.Request.Context(), &param.Product)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"param": param,
			}).Errorf("❌ h.ProductUsecase.CreateNewProduct got an error at %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error_message": err,
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": fmt.Sprintf("Successfully crated new product %d", productID),
		})
		return

	case "edit":
		if param.ID == 0 {
			log.Logger.Error("❌ Params ID is missing")
			c.JSON(http.StatusBadRequest, gin.H{
				"error_message": "Invalid request",
			})
			return
		}

		product, err := h.ProductUsecase.UpdateProduct(c.Request.Context(), &param.Product)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"params": param,
			}).Errorf("❌ h.ProductUsecase.UpdateProduct got an error at %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error_message": err,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message" : "success updating product",
			"product": product,
		})
		return


	case "delete":
		if param.ID == 0 {
			log.Logger.Error("❌ Params ID is missing")
			c.JSON(http.StatusBadRequest, gin.H{
				"error_message": "Invalid request",
			})
			return
		}

		if err := h.ProductUsecase.DeleteProduct(c.Request.Context(), param.ID); err != nil {
			log.Logger.WithFields(logrus.Fields{
				"params": param,
			}).Errorf("❌ h.ProductUsecase.DeleteProduct got an error at %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error_message": err,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Successfully deleted product %s", param.Name),
		})
		return


	default:
		log.Logger.Errorf("❌ Invalid Action %v", param.Action)
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "Invalid action",
		})
		return
	}
}

func (h *ProductHandler) ProductCategoryManagement(c *gin.Context) {
	var param *models.ProductCategoryManagementParameter
	if err := c.ShouldBindJSON(&param); err != nil {
		log.Logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "Invalid input",
		})
		return
	}

	if param.Action == "" {
		log.Logger.Error("❌ Missing Action")
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "Missing parameter action",
		})
		return
	}

	switch param.Action {
	case "add":
		if param.ID != 0 {
			log.Logger.Error("❌ Invalid request, params ID must be empty")
			c.JSON(http.StatusBadRequest, gin.H{
				"error_message": "Invalid request",
			})
			return
		}
		productCategoryID, err := h.ProductUsecase.CreateNewProductCategory(c.Request.Context(), &param.ProductCategory)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"param": param,
			}).Errorf("❌ h.ProductUsecase.CreateNewProductCategory got an error at %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error_message": err,
			})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"message": fmt.Sprintf("Successfully crated new product category %d", productCategoryID),
		})

		return

	case "edit":
		if param.ID == 0 {
			log.Logger.Error("❌ Params ID is missing")
			c.JSON(http.StatusBadRequest, gin.H{
				"error_message": "Invalid request",
			})
			return
		}
		productCategory, err := h.ProductUsecase.UpdateProductCategory(c.Request.Context(), &param.ProductCategory)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"params": param,
			}).Errorf("❌ h.ProductUsecase.UpdateProductCategory got an error at %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error_message": err,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message" : "success updating product",
			"productCategory": productCategory,
		})

		return

	case "delete":
		if param.ID == 0 {
			log.Logger.Error("❌ Params ID is missing")
			c.JSON(http.StatusBadRequest, gin.H{
				"error_message": "Invalid request",
			})
			return
		}

		if err := h.ProductUsecase.DeleteProductCategory(c.Request.Context(), param.ID); err != nil {
			log.Logger.WithFields(logrus.Fields{
				"params": param,
			}).Errorf("❌ h.ProductUsecase.DeleteProductCategory got an error at %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error_message": err,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Successfully deleted product category %d", param.ID),
		})
		
		return

	default:
		log.Logger.Errorf("❌ Invalid Action %v", param.Action)
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "Invalid action",
		})
		return
	}
}

func (h *ProductHandler) GetProductInfo(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"productID": productIDStr,
		}).Errorf("strconv.ParseInt got an error at %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "Invalid product ID",
		})
		return
	}

	product, err := h.ProductUsecase.GetProductByID(c.Request.Context(), productID)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"productID": productID,
		}).Errorf("h.ProductUsecase.GetProductByID got an error at %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error_message": err.Error(),
		})
		return
	}

	if product.ID == 0 {
		log.Logger.WithFields(logrus.Fields{
			"productID": "ProductID not found",
		}).Info("Product ID not found")
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "Product not exist",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"product": product,
	})
}

func (h *ProductHandler) GetProductCategoryInfo(c *gin.Context) {
	productCategoryIDStr := c.Param("id")
	productCategoryID, err := strconv.Atoi(productCategoryIDStr)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"productCategoryID": productCategoryID,
		}).Errorf("strconv.Atoi got an error at %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "Invalid product category ID",
		})
	}

	productCategory, err := h.ProductUsecase.GetProductCategoryByID(c.Request.Context(), productCategoryID)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"productCategoryID": productCategoryID,
		}).Errorf("h.ProductUsecase.GetProductCategoryByID got an error at %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error_message": err.Error(),
		})
		return
	}

	if productCategory.ID == 0 {
		log.Logger.WithFields(logrus.Fields{
			"productCategoryID": productCategoryID,
		}).Info("Product Category ID not found")
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "Product category ID not exist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Product Category": productCategory,
	})
}