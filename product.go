package minicommerce

import (
	"context"
)

// ProductType is the representation of the product type within miniCommerce
type ProductType string

// List of Product types available for a product
const (
	ProductTypeDigital   ProductType = "digital"
	ProductTypeLink      ProductType = "linkable"
	ProductTypeShippable ProductType = "shippable"
)

// Product represents the domain and data model for miniCommerce
type Product struct {
	ID           string            `firestore:"-"`
	Created      int64             `firestore:"created,omitempty"`
	Updated      int64             `firestore:"updated,omitempty"`
	Type         ProductType       `firestore:"type,omitempty"`
	Name         string            `firestore:"name,omitempty"`
	Description  string            `firestore:"description,omitempty"`
	Price        int64             `firestore:"price,omitempty"`
	Metadata     map[string]string `firestore:"metadata,omitempty"`
	Active       bool              `firestore:"active,omitempty"`
	URL          string            `firestore:"url,omitempty"`
	Downloadable []Downloadable    `firestore:"downloadable,omitempty"`
}

// ProductReader is the interface for reading products from a given datastore
type ProductReader interface {
	GetAll(ctx context.Context) ([]Product, error)
	Get(ctx context.Context, id string) (*Product, error)
}

// ProductWriter is the interface for creating a product in a given datastore
type ProductWriter interface {
	Create(ctx context.Context, product *Product) error
}

// ProductUpdater is the interface for updating a product in a given datastor
type ProductUpdater interface {
	Update(ctx context.Context, product *Product) error
}

// ProductRepository is the interface that combines all readers and writers for a product
type ProductRepository interface {
	ProductReader
	ProductWriter
	ProductUpdater
}
