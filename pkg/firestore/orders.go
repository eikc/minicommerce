package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/eikc/minicommerce"
)

// OrdersRepository ...
type OrdersRepository struct {
	client *firestore.Client
}

// NewOrdersRepository ...
func NewOrdersRepository(c *firestore.Client) *OrdersRepository {
	return &OrdersRepository{c}
}

// GetAll ...
func (o *OrdersRepository) GetAll(ctx context.Context) ([]minicommerce.Product, error) {
	return nil, nil
}

// Get ...
func (o *OrdersRepository) Get(ctx context.Context, id string) (*minicommerce.Product, error) {
	return nil, nil
}

// Create ...
func (o *OrdersRepository) Create(ctx context.Context, order *minicommerce.Order) error {
	return nil
}

// Update ...
func (o *OrdersRepository) Update(ctx context.Context, order *minicommerce.Order) error {
	return nil
}
