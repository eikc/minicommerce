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

	defer cleanup(client, ordersCollection, "testing-getting-order")

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

func TestCreateOrder(t *testing.T) {
	ctx := context.Background()
	ID := "testing-order-create"

	c, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		t.Error(err.Error())
	}
	defer cleanup(c, ordersCollection, ID)

	repo := NewOrdersRepository(c)
	order := minicommerce.Order{
		ID:     ID,
		Amount: 15000,
	}

	if err := repo.Create(ctx, &order); err != nil {
		t.Error(err.Error())
	}

	snapshot, err := c.Collection(ordersCollection).Doc(ID).Get(ctx)
	if err != nil {
		t.Error(err.Error())
	}

	cupaloy.SnapshotT(t, snapshot.Data())
}

func TestUpdateOrder(t *testing.T) {
	ctx := context.Background()
	ID := "testing-order-update"

	c, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		t.Error(err.Error())
	}
	defer cleanup(c, ordersCollection, ID)

	orderToUpdate := minicommerce.Order{
		ID:     ID,
		Amount: 15000,
	}

	_, err = c.Collection(ordersCollection).Doc(ID).Create(ctx, orderToUpdate)
	if err != nil {
		t.Error(err.Error())
	}

	repo := NewOrdersRepository(c)
	orderToUpdate.Customer = minicommerce.Customer{
		Name:    "testing update",
		Email:   "testing email",
		Address: "testing address",
		ZipCode: "testing zip code",
		Phone:   "testing phone",
	}

	if err := repo.Update(ctx, &orderToUpdate); err != nil {
		t.Error(err.Error())
	}

	snapshot, err := c.Collection(ordersCollection).Doc(ID).Get(ctx)
	if err != nil {
		t.Error(err.Error())
	}

	cupaloy.SnapshotT(t, snapshot.Data())
}
