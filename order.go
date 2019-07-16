package minicommerce

// Order represents the domain model for an order or cart in minicommerce
type Order struct {
	ID        string    `firestore:"-"`
	PaymentID string    `firestore:"paymentId"`
	Coupon    string    `firestore:"coupon"`
	Items     []Product `firestore:"items"`
	Customer  Customer  `firestore:"customer"`
	Refunded  bool      `firestore:"refunded"`
	Amount    int64     `firestore:"amount"`
	Discount  int64     `firestore:"discount"`
	Shipping  int64     `firestore:"shipping"`
	NetAmount int64     `firestore:"netAmount"`
	Taxes     int64     `firestore:"taxes"`
	Total     int64     `firestore:"total"`
}

// Customer is...
type Customer struct {
	Name    string `firestore:"name"`
	Email   string `firestore:"email"`
	Address string `firestore:"address"`
	ZipCode string `firestore:"zipCode"`
	Phone   string `firestore:"phone"`
}
