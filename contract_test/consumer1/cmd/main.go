package main

import (
	"log"
	"net/url"

	"consumer1"
)

func main() {
	u, _ := url.Parse("http://localhost:8080")
	client := &consumer1.Client{
		BaseURL: u,
	}

	products, err := client.GetProducts()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Products:")
	log.Println(products)

	product, err := client.GetProduct(1)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Product with ID 1:")
	log.Println(product)
}
