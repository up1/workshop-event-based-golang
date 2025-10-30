package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"provider1"
	_ "provider1/docs" // This line is needed for swagger to find the generated docs
)

// @title Product Service API
// @version 1.0
// @description This is a sample product service server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
func main() {
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/products", provider1.GetProducts).Methods("GET")
	api.HandleFunc("/products/{id}", provider1.GetProduct).Methods("GET")

	// Swagger
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	log.Println("Server starting on :8080")
	log.Println("Swagger UI available at: http://localhost:8080/swagger/index.html")
	log.Fatal(http.ListenAndServe(":8080", r))
}
