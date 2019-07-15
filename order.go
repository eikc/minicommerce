package minicommerce

// Order represents the domain model for an order or cart in minicommerce
type Order struct {
	ID        string    `firestore:"-"`
	PaymentID string    `firestore:"paymentId,omitempty"`
	Coupon    string    `firestore:"coupon,omitempty"`
	Items     []Product `firestore:"items,omitempty"`
	Customer  Customer  `firestore:"customer,omitempty"`
	Refunded  bool      `firestore:"refunded,omitempty"`
	Amount    int64     `firestore:"amount,omitempty"`
	Discount  int64     `firestore:"discount,omitempty"`
	Shipping  int64     `firestore:"shipping,omitempty"`
	NetAmount int64     `firestore:"netAmount,omitempty"`
	Taxes     int64     `firestore:"taxes,omitempty"`
	Total     int64     `firestore:"total,omitempty"`
}

// Customer is...
type Customer struct {
	Name    string `firestore:"name,omitempty"`
	Email   string `firestore:"email,omitempty"`
	Address string `firestore:"address,omitempty"`
	ZipCode string `firestore:"zipCode,omitempty"`
	Phone   string `firestore:"phone,omitempty"`
}
