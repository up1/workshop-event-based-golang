package model

import (
	"errors"
)

// User represents a product in the system
type Product struct {
	ProductName string `json:"productName" pact:"example=Product A"`
	Price       int    `json:"price" pact:"example=1999"`
	Stock       int    `json:"stock" pact:"example=50"`
	ID          int    `json:"id" pact:"example=10"`
}

var (
	// ErrNotFound represents a resource not found (404)
	ErrNotFound = errors.New("not found")

	// ErrEmpty is returned when input string is empty
	ErrEmpty = errors.New("empty string")
)

// ProductResponse represents the response structure for a product
type ProductResponse struct {
	Product *Product `json:"product"`
}
