// Package service holds product business logic.
package service

import "archview-example-gin-mvc/internal/product/repository"

// ProductService is the product use-case boundary (single impl: productService).
type ProductService interface {
	ListProducts() []repository.Product
	CreateProduct(name string) repository.Product
}

type productService struct {
	repo repository.ProductRepository
}

// New wires a ProductService over a ProductRepository.
func New(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) ListProducts() []repository.Product {
	return s.repo.FindAllProducts()
}

func (s *productService) CreateProduct(name string) repository.Product {
	return s.repo.InsertProduct(repository.Product{Name: name})
}
