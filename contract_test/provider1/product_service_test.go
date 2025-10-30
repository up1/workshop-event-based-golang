package provider1_test

import (
	"encoding/json"
	"model"
	"net/http"
	"net/http/httptest"
	"provider1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProducts(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/products", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(provider1.GetProducts)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var products []*model.Product
	err = json.Unmarshal(rr.Body.Bytes(), &products)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, products, 2)
	assert.Equal(t, "Product 1", products[0].ProductName)
	assert.Equal(t, 100, products[0].Price)
	assert.Equal(t, 10, products[0].Stock)
	assert.Equal(t, 1, products[0].ID)

	assert.Equal(t, "Product 2", products[1].ProductName)
	assert.Equal(t, 200, products[1].Price)
	assert.Equal(t, 20, products[1].Stock)
	assert.Equal(t, 2, products[1].ID)
}

func TestGetProductByID(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/products/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(provider1.GetProduct)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var product model.Product
	err = json.Unmarshal(rr.Body.Bytes(), &product)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "Product 1", product.ProductName)
	assert.Equal(t, 100, product.Price)
	assert.Equal(t, 10, product.Stock)
	assert.Equal(t, 1, product.ID)
}

func TestGetProductByIDNotFound(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/products/999", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(provider1.GetProduct)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
