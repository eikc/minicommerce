package minicommerce

// Coupon is the domain and data model representing a Coupon in miniCommerce
type Coupon struct {
	ID             string  `firestore:"-"`
	Description    string  `firestore:"description"`
	Active         bool    `firestore:"active"`
	AmountOff      int64   `firestore:"amountOff"`
	PercentOff     float64 `firestore:"percentOff"`
	MaxRedemptions int64   `firestore:"maxRedemptions"`
	RedeemBy       int64   `firestore:"redeemBy"`
	RedeemBefore   int64   `firestore:"redeemBefore"`
}
