package firestore

import (
	"context"
	"fmt"

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
	docRef := o.client.Collection(ordersCollection).Doc(id)
	snapshot, err := docRef.Get(ctx)
	if err != nil {
		return nil, err
	}

	if !snapshot.Exists() {
		return nil, &DocumentNotFoundError{fmt.Sprintf("%s/%s", ordersCollection, id)}
	}

	order := minicommerce.Order{
		ID: id,
	}

	if err := snapshot.DataTo(&order); err != nil {
		return nil, err
	}

	return &order, nil
}

// Create ...
func (o *OrdersRepository) Create(ctx context.Context, order *minicommerce.Order) error {
	docRef := o.client.Collection(ordersCollection).Doc(order.ID)
	if _, err := docRef.Create(ctx, order); err != nil {
		return err
	}

	return nil
}

// Update updates the existing orders document by replacing it using the firestore set method
func (o *OrdersRepository) Update(ctx context.Context, order *minicommerce.Order) error {
	docRef := o.client.Collection(ordersCollection).Doc(order.ID)
	if _, err := docRef.Set(ctx, order); err != nil {
		return err
	}

	return nil
}
