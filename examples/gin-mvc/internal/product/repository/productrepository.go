// Package repository is the product data layer.
package repository

// Product is a demo product record.
type Product struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ProductRepository abstracts product persistence (single impl: productRepository).
type ProductRepository interface {
	FindAllProducts() []Product
	InsertProduct(p Product) Product
}

type productRepository struct {
	rows []Product
	next int
}

// New returns an in-memory ProductRepository.
func New() ProductRepository {
	return &productRepository{
		rows: []Product{{ID: 1, Name: "Keyboard"}, {ID: 2, Name: "Monitor"}},
		next: 3,
	}
}

func (r *productRepository) FindAllProducts() []Product { return r.rows }

func (r *productRepository) InsertProduct(p Product) Product {
	p.ID = r.next
	r.next++
	r.rows = append(r.rows, p)
	return p
}
