package minicommerce

// Coupon is the domain and data model representing a Coupon in miniCommerce
type Coupon struct {
	ID             string  `firestore:"id,omitempty"`
	Name           string  `firestore:"name,omitempty"`
	AmountOff      int64   `firestore:"amountOff,omitempty"`
	PercentOff     float64 `firestore:"percentOff,omitempty"`
	MaxRedemptions int64   `firestore:"maxRedemptions,omitempty"`
	RedeemBy       int64   `firestore:"redeemBy,omitempty"`
	RedeemBefore   int64   `firestore:"redeemBefore,omitempty"`
}
