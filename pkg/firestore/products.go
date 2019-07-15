package firestore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/eikc/minicommerce"
)

const productsCollection string = "products"

// ProductRepository is the struct that handle all communication with firestore when working with products
type ProductRepository struct {
	client *firestore.Client
}

// NewProductRepository is a constructor helper for ProductRepository
func NewProductRepository(c *firestore.Client) *ProductRepository {
	return &ProductRepository{c}
}

// GetAll ...
func (p *ProductRepository) GetAll(ctx context.Context) ([]minicommerce.Product, error) {
	colRef := p.client.Collection(productsCollection)
	iter := colRef.Documents(ctx)
	docs, err := iter.GetAll()
	if err != nil {
		return nil, err
	}

	var products []minicommerce.Product
	for _, p := range docs {
		product := minicommerce.Product{
			ID: p.Ref.ID,
		}

		if err := p.DataTo(&product); err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	return products, nil
}

// Get ...
func (p *ProductRepository) Get(ctx context.Context, id string) (*minicommerce.Product, error) {
	docRef := p.client.Collection(productsCollection).Doc(id)
	snapshot, err := docRef.Get(ctx)
	if err != nil {
		return nil, err
	}

	if !snapshot.Exists() {
		return nil, &DocumentNotFoundError{fmt.Sprintf("%s/%s", downloadableCollection, id)}
	}

	product := &minicommerce.Product{
		ID: id,
	}

	if err := snapshot.DataTo(product); err != nil {
		return nil, err
	}

	return product, nil
}

// Create ...
func (p *ProductRepository) Create(ctx context.Context, product *minicommerce.Product) error {
	docRef := p.client.Collection(productsCollection).Doc(product.ID)
	if _, err := docRef.Create(ctx, product); err != nil {
		return err
	}

	return nil
}

// Update ...
func (p *ProductRepository) Update(ctx context.Context, product *minicommerce.Product) error {
	docRef := p.client.Collection(productsCollection).Doc(product.ID)
	if _, err := docRef.Set(ctx, product); err != nil {
		return err
	}

	return nil
}
