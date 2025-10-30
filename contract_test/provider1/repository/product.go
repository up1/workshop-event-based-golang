package repository

import "model"

// ProductRepository is an in-memory db representation of our set of products
type ProductRepository struct {
	Products map[string]*model.Product
}

// GetProducts returns all products in the repository
func (p *ProductRepository) GetProducts() []model.Product {
	var response []model.Product

	for _, product := range p.Products {
		response = append(response, *product)
	}

	return response
}

// ByID finds a product by their ID
func (p *ProductRepository) ByID(ID int) (*model.Product, error) {
	for _, product := range p.Products {
		if product.ID == ID {
			return product, nil
		}
	}
	return nil, model.ErrNotFound
}
