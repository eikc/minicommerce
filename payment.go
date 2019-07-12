package minicommerce

// Payment represents the domain model for payments within the system
type Payment struct {
	ID         string `firestore:"id,omitempty"`
	ExternalID string `firestore:"externalID,omitempty"`
	Amount     int64  `firestore:"amount,omitempty"`
	Paid       bool   `firestore:"paid,omitempty"`
	Refunded   bool   `firestore:"refunded,omitempty"`
}
