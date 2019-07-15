package firestore

import (
	"context"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"

	"github.com/eikc/minicommerce"

	"cloud.google.com/go/firestore"
)

func TestGetAllCoupons(t *testing.T) {
	ctx := context.Background()
	cc := []minicommerce.Coupon{
		{
			ID:             "coupon-get-all-1",
			Description:    "test coupon",
			Active:         true,
			AmountOff:      10000,
			PercentOff:     0.10,
			MaxRedemptions: 5,
			RedeemBy:       0,
			RedeemBefore:   0,
		},
		{
			ID:             "coupon-get-all-2",
			Description:    "test coupon",
			Active:         false,
			AmountOff:      5000,
			PercentOff:     0.10,
			MaxRedemptions: 5,
			RedeemBy:       0,
			RedeemBefore:   0,
		},
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		t.Error(err.Error())
	}

	defer func() {
		for _, c := range cc {
			client.Collection(couponsCollection).Doc(c.ID).Delete(ctx)
		}
		client.Close()
	}()

	for _, c := range cc {
		if _, err := client.Collection(couponsCollection).Doc(c.ID).Set(ctx, c); err != nil {
			t.Errorf(err.Error())
		}
	}

	repo := NewCouponsRepository(client)

	coupons, err := repo.GetAll(ctx)
	if err != nil {
		t.Error(err.Error())
	}

	cupaloy.SnapshotT(t, coupons)
}

func TestGetCouponByCode(t *testing.T) {
	ctx := context.Background()
	c := minicommerce.Coupon{
		ID:             "get-by-code",
		Description:    "Trying to get a coupon by code",
		Active:         true,
		AmountOff:      500,
		PercentOff:     0.10,
		MaxRedemptions: 10,
		RedeemBy:       2,
		RedeemBefore:   1563198147,
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		t.Error(err.Error())
	}

	defer func() {
		client.Collection(couponsCollection).Doc(c.ID).Delete(ctx)
		client.Close()
	}()

	_, err = client.Collection(couponsCollection).Doc(c.ID).Set(ctx, c)
	if err != nil {
		t.Error(err.Error())
	}

	repo := NewCouponsRepository(client)
	coupon, err := repo.GetByCode(ctx, "get-by-code")
	if err != nil {
		t.Error(err.Error())
	}

	cupaloy.SnapshotT(t, coupon)
}

func TestUpdateCoupon(t *testing.T) {
	ctx := context.Background()
	c := minicommerce.Coupon{
		ID:             "update-code",
		Description:    "Trying to get a coupon by code",
		Active:         true,
		AmountOff:      500,
		PercentOff:     0.10,
		MaxRedemptions: 10,
		RedeemBy:       2,
		RedeemBefore:   1563198147,
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		t.Error(err.Error())
	}

	defer func() {
		client.Collection(couponsCollection).Doc(c.ID).Delete(ctx)
		client.Close()
	}()

	_, err = client.Collection(couponsCollection).Doc(c.ID).Set(ctx, c)
	if err != nil {
		t.Error(err.Error())
	}

	repo := NewCouponsRepository(client)

	c.Active = false
	c.RedeemBy++

	if err := repo.Update(ctx, c); err != nil {
		t.Error(err.Error())
	}

	docRef := client.Collection(couponsCollection).Doc(c.ID)
	snapshot, err := docRef.Get(ctx)
	if err != nil {
		t.Error(err.Error())
	}

	cupaloy.SnapshotT(t, snapshot.Data())
}
