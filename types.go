package main

const (
	flowStatus            = "flowstatus"
	customerCreatedStatus = "UserCreated"
	invoiceCreatedStatus  = "InvoiceCreated"
	invoiceBookedStatus   = "invoiceBooked"
	invoicePaidStatus     = "InvoicePaid"
	invoiceSentStatus     = "EmailSent"
)

type Settings struct {
	ClientKey              string
	ClientSecret           string
	APIKey                 string
	OrganizationID         int
	StripeKey              string
	StripeWebhookSignature string
	InstagramToken         string
}

type Bootcamp struct {
	ID        string
	Date      string
	Location  string
	StartsAt  string
	SpotsLeft int64
}
