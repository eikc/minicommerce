package main

import (
	"fmt"

	"github.com/eikc/dinero-go"
	"github.com/eikc/dinero-go/contacts"
	"github.com/eikc/dinero-go/invoices"
	"github.com/eikc/dinero-go/ledgeritems"
)

type dineroAPI struct {
	dinero.API
}

type invoiceCreated struct {
	ID        string
	Number    int
	Timestamp string
}

func (api *dineroAPI) CreateCustomer(email, name, address string) (string, error) {
	cParams := contacts.ContactParams{
		IsPerson:             true,
		Name:                 name,
		Email:                email,
		Street:               address,
		CountryKey:           "DK",
		PaymentConditionType: dinero.NettoCash,
	}

	c, err := contacts.Add(api, cParams)
	if err != nil {
		return "", err
	}

	return c.ID, nil
}

func (api *dineroAPI) CreateInvoice(customerID, description string, amount int64) (*invoiceCreated, error) {
	invoiceParams := invoices.CreateInvoice{
		ContactID:        customerID,
		ShowLinesInclVat: true,
		Currency:         "DKK",
		Language:         "da-DK",
		Date:             dinero.DateNow(),
		ProductLines: []invoices.InvoiceLine{
			invoices.InvoiceLine{
				BaseAmountValue: float64((amount / 100)),
				Quantity:        1,
				AccountNumber:   1000,
				Description:     description,
				LineType:        "Product",
				Unit:            "parts",
			},
		},
	}

	timestamp, err := invoices.Save(api, invoiceParams)
	if err != nil {
		return nil, err
	}

	invoice, err := invoices.Get(api, timestamp.ID)
	if err != nil {
		return nil, err
	}

	return &invoiceCreated{invoice.ID, invoice.Number, invoice.Timestamp}, nil
}

func (api *dineroAPI) BookInvoice(invoiceID, timestamp string) (*invoiceCreated, error) {
	invoice, err := invoices.Get(api, invoiceID)
	if err != nil {
		return nil, err
	}

	var isBooked = invoice.Status == "Booked"

	if !isBooked {
		_, err := invoices.Book(api, invoiceID, timestamp)
		if err != nil {
			return nil, err
		}

		invoice, err = invoices.Get(api, invoiceID)
		if err != nil {
			return nil, err
		}
	}

	return &invoiceCreated{invoice.ID, invoice.Number, invoice.Timestamp}, nil
}

func (api *dineroAPI) CreatePayment(invoiceID string, amount int64) error {
	invoice, err := invoices.Get(api, invoiceID)
	if err != nil {
		return err
	}

	isPaid := invoice.PaymentStatus == "Paid"
	if isPaid {
		return nil
	}

	payment := invoices.CreatePaymentParams{
		Amount:               float64(amount / 100),
		DepositAccountNumber: 55010,
		Description:          "Paid with stripe",
		PaymentDate:          dinero.DateNow(),
		Timestamp:            invoice.Timestamp,
	}

	_, err = invoices.CreatePayment(api, invoiceID, payment)
	if err != nil {
		return err
	}

	return nil
}

func (api *dineroAPI) SendInvoice(invoiceID string) error {
	_, err := invoices.SendEmail(api, invoiceID, invoices.SendInvoice{})
	if err != nil {
		return err
	}

	return nil
}

func (api *dineroAPI) AddStripePayout(payoutID string, amount, fee float64) error {

	ledgerItems := []ledgeritems.LedgerItem{
		{
			AccountNumber:  55000,
			AccountVatCode: "None",
			Amount:         amount,
			Description:    fmt.Sprintf("Stripe payout udbetaling: %s", payoutID),
			VoucherNumber:  1,
			VoucherDate:    dinero.DateNow(),
		},
		{
			AccountNumber:  7220,
			AccountVatCode: "None",
			Amount:         fee,
			Description:    fmt.Sprintf("Stripe gebyr, payout: %s", payoutID),
			VoucherNumber:  1,
			VoucherDate:    dinero.DateNow(),
		},
		{
			AccountNumber:  55010,
			AccountVatCode: "None",
			Amount:         (amount * -1) + (fee * -1),
			Description:    fmt.Sprintf("Stripe payout udbetaling: %s", payoutID),
			VoucherNumber:  1,
			VoucherDate:    dinero.DateNow(),
		},
	}

	_, err := ledgeritems.Create(api, ledgerItems)

	return err
}
