package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/eikc/minicommerce"
)

const ordersCollection string = "orders"

// OrdersRepository ...
type OrdersRepository struct {
	client *firestore.Client
}

// NewOrdersRepository ...
func NewOrdersRepository(c *firestore.Client) *OrdersRepository {
	return &OrdersRepository{c}
}

// GetAll ...
func (o *OrdersRepository) GetAll(ctx context.Context) ([]minicommerce.Order, error) {
	colRef := o.client.Collection(ordersCollection)
	iter := colRef.Documents(ctx)
	docs, err := iter.GetAll()
	if err != nil {
		return nil, err
	}

	var orders []minicommerce.Order

	for _, d := range docs {
		order := minicommerce.Order{
			ID: d.Ref.ID,
		}

		if err := d.DataTo(&order); err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	return orders, nil
}

// Get ...
func (o *OrdersRepository) Get(ctx context.Context, id string) (*minicommerce.Order, error) {
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
