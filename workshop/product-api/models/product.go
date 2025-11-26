// models/product.go
package models

// Product represents the product detail.
type Product struct {
	ID    int     `json:"id" example:"1"`
	Name  string  `json:"name" example:"Product name 1"`
	Price float64 `json:"price" example:"100.50"`
	Stock int     `json:"stock" example:"10"`
}

// ErrorResponse represents the response for an error.
type ErrorResponse struct {
	Message string `json:"message" example:"product id=2 not found in system"`
}
