package provider1

import (
	"encoding/json"
	"net/http"
	"provider1/model"
	"provider1/repository"
	"strconv"

	"github.com/gorilla/mux"
)

// productRepository is a mock in-memory representation of our product repository
var productRepository = &repository.ProductRepository{
	Products: map[string]*model.Product{
		"product1": {
			ProductName: "Product 1",
			Price:       100,
			Stock:       10,
			ID:          1,
		},
		"product2": {
			ProductName: "Product 2",
			Price:       200,
			Stock:       20,
			ID:          2,
		},
	},
}

// GetProducts handles the HTTP request to retrieve all products
// @Summary Get all products
// @Description Get all products from the system
// @Tags products
// @Accept json
// @Produce json
// @Success 200 {array} model.Product
// @Router /products [get]
func GetProducts(w http.ResponseWriter, r *http.Request) {
	products := productRepository.GetProducts()
	w.Header().Set("Content-Type", "application/json")
	resBody, _ := json.Marshal(products)
	w.Write(resBody)
}

// GetProduct handles the HTTP request to retrieve a product by its ID
// @Summary Get a product by ID
// @Description Get a single product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} model.Product
// @Failure 404 {object} map[string]string
// @Router /products/{id} [get]
func GetProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get product ID from mux vars
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	product, err := productRepository.ByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Product not found"})
	} else {
		w.WriteHeader(http.StatusOK)
		resBody, _ := json.Marshal(product)
		w.Write(resBody)
	}
}
