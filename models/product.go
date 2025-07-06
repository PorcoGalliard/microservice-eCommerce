package models

type (
	Product struct {
		ID int64 `json:"id"`
		Name string `json:"name"`
		Description string `json:"description"`
		Price float64 `json:"price"`
		Stock int `json:"stock"`
		Category_ID int64 `json:"category_id"`
	}

	ProductCategory struct {
		ID int `json:"id"`
		Name string `json:"name"`
	}

	ProductCategoryManagementParameter struct {
		Action string `json:"action"`
		ProductCategory
	}

	ProductManagementParameter struct {
		Action string `json:"action"`
		Product
	}
)