package firestore

import (
	"context"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"

	"github.com/eikc/minicommerce"

	"cloud.google.com/go/firestore"
)

func TestGetProduct(t *testing.T) {
	ctx := context.Background()
	ID := "testing-get-product"

	c, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		t.Errorf(err.Error())
	}

	defer func() {
		c.Collection(productsCollection).Doc(ID).Delete(ctx)
		c.Close()
	}()

	p := minicommerce.Product{
		ID:          ID,
		Created:     1,
		Updated:     1,
		Type:        minicommerce.ProductTypeDigital,
		Name:        "One digital product",
		Description: "And it has a description",
		Price:       10000,
		Active:      true,
		Downloadable: []minicommerce.Downloadable{
			{ID: "testing", Name: "One digital product", Location: "foodie.pdf"},
		},
	}

	if _, err := c.Collection(productsCollection).Doc(ID).Set(ctx, p); err != nil {
		t.Errorf(err.Error())
	}

	repo := NewProductRepository(c)

	result, err := repo.Get(ctx, ID)
	if err != nil {
		t.Errorf(err.Error())
	}

	cupaloy.SnapshotT(t, result)
}

func TestGetAllProducts(t *testing.T) {
	ctx := context.Background()
	c, err := firestore.NewClient(ctx, projectID)

	if err != nil {
		t.Errorf(err.Error())
	}

	pp := []minicommerce.Product{
		minicommerce.Product{
			ID:          "product-one",
			Created:     1,
			Updated:     1,
			Type:        minicommerce.ProductTypeDigital,
			Name:        "One digital product",
			Description: "And it has a description",
			Price:       10000,
			Active:      true,
			Downloadable: []minicommerce.Downloadable{
				{ID: "testing", Name: "One digital product", Location: "foodie.pdf"},
			},
		},
		minicommerce.Product{
			ID:          "product-two",
			Created:     1,
			Updated:     1,
			Type:        minicommerce.ProductTypeDigital,
			Name:        "One digital product",
			Description: "And it has a description",
			Price:       10000,
			Active:      true,
			Downloadable: []minicommerce.Downloadable{
				{ID: "testing", Name: "One digital product", Location: "foodie.pdf"},
			},
		},
		minicommerce.Product{
			ID:          "product-three",
			Created:     1,
			Updated:     1,
			Type:        minicommerce.ProductTypeDigital,
			Name:        "One digital product",
			Description: "And it has a description",
			Price:       10000,
			Active:      true,
			Downloadable: []minicommerce.Downloadable{
				{ID: "testing", Name: "One digital product", Location: "foodie.pdf"},
			},
		},
	}

	defer func() {
		for _, p := range pp {
			c.Collection(productsCollection).Doc(p.ID).Delete(ctx)
		}
		c.Close()
	}()

	for _, p := range pp {
		docRef := c.Collection(productsCollection).Doc(p.ID)
		if _, err := docRef.Set(ctx, p); err != nil {
			t.Errorf(err.Error())
		}
	}

	repo := NewProductRepository(c)

	docs, err := repo.GetAll(ctx)
	if err != nil {
		t.Errorf(err.Error())
	}

	cupaloy.SnapshotT(t, docs)
}

func TestCreateProduct(t *testing.T) {
	ctx := context.Background()
	ID := "testing-product-create"
	c, err := firestore.NewClient(ctx, projectID)

	if err != nil {
		t.Errorf(err.Error())
	}

	defer func() {
		c.Collection(productsCollection).Doc(ID).Delete(ctx)
		c.Close()
	}()

	repo := NewProductRepository(c)

	p := minicommerce.Product{
		ID:          ID,
		Created:     1,
		Updated:     2,
		Type:        minicommerce.ProductTypeShippable,
		Name:        "testing create product",
		Description: "with a description",
		Price:       10000,
	}

	if err := repo.Create(ctx, &p); err != nil {
		t.Errorf(err.Error())
	}
}

func TestUpdateProduct(t *testing.T) {
	ctx := context.Background()
	ID := "testing-product-update"

	c, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		t.Errorf(err.Error())
	}

	defer func() {
		c.Collection(productsCollection).Doc(ID).Delete(ctx)
		c.Close()
	}()

	p := minicommerce.Product{
		ID:      ID,
		Created: 1,
		Updated: 2,
		Name:    "first name",
	}

	if _, err := c.Collection(productsCollection).Doc(ID).Set(ctx, p); err != nil {
		t.Errorf(err.Error())
	}

	p.Name = "new name"
	p.Updated = 3

	repo := NewProductRepository(c)

	if err := repo.Update(ctx, &p); err != nil {
		t.Errorf(err.Error())
	}
}
