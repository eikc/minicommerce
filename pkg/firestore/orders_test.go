package firestore

import (
	"context"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"

	"github.com/eikc/minicommerce"

	"cloud.google.com/go/firestore"
)

func TestGetAllOrders(t *testing.T) {
	ctx := context.Background()
	oo := []minicommerce.Order{
		{
			ID:        "get-all-orders-1",
			Amount:    100,
			Shipping:  25,
			NetAmount: 125,
			Discount:  0,
			Taxes:     25,
			Total:     150,
		},
		{
			ID: "get-all-orders-2",
		},
		{
			ID: "get-all-orders-3",
		},
	}

	c, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		t.Error(err.Error())
	}

	defer func() {
		for _, o := range oo {
			c.Collection(ordersCollection).Doc(o.ID).Delete(ctx)
		}
		c.Close()
	}()

	for _, o := range oo {
		doc := c.Collection(ordersCollection).Doc(o.ID)
		if _, err := doc.Set(ctx, o); err != nil {
			t.Error(err.Error())
		}
	}

	repo := NewOrdersRepository(c)
	orders, err := repo.GetAll(ctx)
	if err != nil {
		t.Error(err.Error())
	}

	cupaloy.SnapshotT(t, orders)
}
