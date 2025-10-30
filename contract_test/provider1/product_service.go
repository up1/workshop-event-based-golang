package provider1

import (
	"encoding/json"
	"net/http"
	"provider1/model"
	"provider1/repository"
	"strconv"
	"strings"
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
func GetProducts(w http.ResponseWriter, r *http.Request) {
	products := productRepository.GetProducts()
	w.Header().Set("Content-Type", "application/json")
	resBody, _ := json.Marshal(products)
	w.Write(resBody)
}

// GetProduct handles the HTTP request to retrieve a product by its ID
func GetProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get product ID from path
	a := strings.Split(r.URL.Path, "/")
	id, _ := strconv.Atoi(a[len(a)-1])

	product, err := productRepository.ByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
		resBody, _ := json.Marshal(product)
		w.Write(resBody)
	}
}
