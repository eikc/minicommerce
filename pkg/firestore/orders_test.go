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

func TestGetOrder(t *testing.T) {
	ctx := context.Background()
	o := minicommerce.Order{
		ID:        "testing-getting-order",
		PaymentID: "payment-intent",
		Coupon:    "testing-coupon",
		Items: []minicommerce.Product{
			{
				ID:          "product-id",
				Created:     1,
				Updated:     2,
				Type:        minicommerce.ProductTypeDigital,
				Name:        "det lille skridt",
				Description: "en bog om madsens skridt",
				Price:       15000,
				Active:      true,
				Downloadable: []minicommerce.Downloadable{
					{
						ID:       "testing-downlodable",
						Name:     "det-lille-skridt.pdf",
						Location: "det-lille-skridt.pdf",
					},
				},
			},
		},
		Customer: minicommerce.Customer{
			Name:    "testing name",
			Email:   "testing email",
			Address: "testing address",
			ZipCode: "zipcode testing",
			Phone:   "phone field",
		},
		Refunded:  false,
		Amount:    15000,
		Discount:  0,
		Shipping:  0,
		NetAmount: 15000,
		Taxes:     5000,
		Total:     20000,
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		t.Error(err.Error())
	}

	if _, err := client.Collection(ordersCollection).Doc(o.ID).Set(ctx, o); err != nil {
		t.Error(err.Error())
	}

	repo := NewOrdersRepository(client)

	order, err := repo.Get(ctx, o.ID)
	if err != nil {
		t.Error(err.Error())
	}

	cupaloy.SnapshotT(t, order)
}
