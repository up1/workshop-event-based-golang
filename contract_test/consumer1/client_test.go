package consumer1_test

import (
	"consumer1"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"model"

	"github.com/stretchr/testify/assert"
)

func TestClientUnit_GetProduct(t *testing.T) {
	productID := 1

	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.URL.String(), fmt.Sprintf("/api/v1/products/%d", productID))
		product, _ := json.Marshal(model.Product{
			ID:          productID,
			ProductName: "Test Product",
			Price:       150,
			Stock:       30,
		})
		rw.Write([]byte(product))
	}))
	defer server.Close()

	// Setup client
	u, _ := url.Parse(server.URL)
	client := &consumer1.Client{
		BaseURL: u,
	}

	// Act
	product, err := client.GetProduct(productID)
	assert.NoError(t, err)

	// Assert
	assert.Equal(t, product.ID, productID)
	assert.Equal(t, product.ProductName, "Test Product")
	assert.Equal(t, product.Price, 150)
	assert.Equal(t, product.Stock, 30)
}

func TestClientUnit_GetProducts(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.URL.String(), "/api/v1/products")
		products, _ := json.Marshal([]model.Product{
			{
				ID:          1,
				ProductName: "Test Product 1",
				Price:       150,
				Stock:       30,
			},
			{
				ID:          2,
				ProductName: "Test Product 2",
				Price:       250,
				Stock:       20,
			},
		})
		rw.Write([]byte(products))
	}))
	defer server.Close()

	// Setup client
	u, _ := url.Parse(server.URL)
	client := &consumer1.Client{
		BaseURL: u,
	}

	// Act
	products, err := client.GetProducts()
	assert.NoError(t, err)

	// Assert
	assert.Len(t, products, 2)

	assert.Equal(t, products[0].ID, 1)
	assert.Equal(t, products[0].ProductName, "Test Product 1")
	assert.Equal(t, products[0].Price, 150)
	assert.Equal(t, products[0].Stock, 30)

	assert.Equal(t, products[1].ID, 2)
	assert.Equal(t, products[1].ProductName, "Test Product 2")
	assert.Equal(t, products[1].Price, 250)
	assert.Equal(t, products[1].Stock, 20)
}
