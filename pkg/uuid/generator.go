package uuid

import (
	"github.com/gofrs/uuid"
)

// Generator struct knows how to generate ID's in minicommerce
// it is the recommened way to generate ID's
type Generator struct {
}

// NewGenerator is the constructor function for a Generator
func NewGenerator() *Generator {
	return &Generator{}
}

// New generates a new UUID v4 ID to use for models
func (g *Generator) New() (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	return id.String(), nil
}
