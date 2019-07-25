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
	ID           string            `firestore:"-" json:"id"`
	Created      int64             `firestore:"created" json:"created"`
	Updated      int64             `firestore:"updated" json:"updated"`
	Type         ProductType       `firestore:"type" json:"type"`
	Name         string            `firestore:"name" json:"name"`
	Description  string            `firestore:"description" json:"description"`
	Price        int64             `firestore:"price" json:"price"`
	Metadata     map[string]string `firestore:"metadata" json:"metadata"`
	Active       bool              `firestore:"active" json:"active"`
	URL          string            `firestore:"url" json:"url"`
	Downloadable []Downloadable    `firestore:"downloadable" json:"downloadables"`
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
