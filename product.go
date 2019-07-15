package minicommerce

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
