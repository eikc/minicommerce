package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/eikc/minicommerce"
)

const couponsCollection string = "coupons"

// CouponsRepository is the repository that communicates with the firestore database when handling coupon codes
type CouponsRepository struct {
	client *firestore.Client
}

// NewCouponsRepository constructs the coupons repository
func NewCouponsRepository(c *firestore.Client) *CouponsRepository {
	return &CouponsRepository{c}
}

// GetAll ...
func (c *CouponsRepository) GetAll(ctx context.Context) ([]minicommerce.Coupon, error) {
	colRef := c.client.Collection(couponsCollection)
	iter := colRef.Documents(ctx)

	docs, err := iter.GetAll()
	if err != nil {
		return nil, err
	}

	var coupons []minicommerce.Coupon

	for _, d := range docs {
		c := minicommerce.Coupon{
			ID: d.Ref.ID,
		}
		if err := d.DataTo(&c); err != nil {
			return nil, err
		}
		coupons = append(coupons, c)
	}

	return coupons, nil
}

// GetByCode ...
func (c *CouponsRepository) GetByCode(ctx context.Context, code string) (*minicommerce.Coupon, error) {
	docRef := c.client.Collection(couponsCollection).Doc(code)
	snapshot, err := docRef.Get(ctx)
	if err != nil {
		return nil, err
	}

	coupon := minicommerce.Coupon{
		ID: code,
	}

	if err := snapshot.DataTo(&coupon); err != nil {
		return nil, err
	}

	return &coupon, nil
}

// Create ...
func (c *CouponsRepository) Create(ctx context.Context, coupon minicommerce.Coupon) error {
	docRef := c.client.Collection(couponsCollection).Doc(coupon.ID)
	_, err := docRef.Create(ctx, coupon)
	if err != nil {
		return err
	}

	return nil
}

// Update ...
func (c *CouponsRepository) Update(ctx context.Context, coupon minicommerce.Coupon) error {
	docRef := c.client.Collection(couponsCollection).Doc(coupon.ID)
	_, err := docRef.Set(ctx, coupon)
	if err != nil {
		return err
	}

	return nil
}
