package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/stripe/stripe-go"
	stripeClient "github.com/stripe/stripe-go/client"
)

// ProgramWorkFlow handles all actions surrounding the workflow of a program
type ProgramWorkFlow struct {
	DineroAPI  *dineroAPI
	StripeAPI  *stripeClient.API
	httpClient *http.Client
}

// StartFlow Controls the workflow of the order
func (workflow *ProgramWorkFlow) StartFlow(o stripe.Order) (string, error) {
	flow := o.Metadata[flowStatus]
	var err error

	switch flow {
	case customerCreatedStatus:
		err = workflow.CreateInvoice(o)
	case invoiceCreatedStatus:
		err = workflow.BookInvoice(o)
	case invoiceBookedStatus:
		err = workflow.CreatePayment(o)
	case invoicePaidStatus:
		err = workflow.FulfillWorkflow(o)
	case invoiceSentStatus:
		workflow.CelebrateOrder(o)
	}

	return flow, err
}

// CreateCustomer Creates a customer in the given workflow
func (workflow *ProgramWorkFlow) CreateCustomer(o stripe.Order) error {
	name := o.Metadata["name"]
	email := o.Metadata["email"]
	address := o.Metadata["address"]
	contactID, err := workflow.DineroAPI.CreateCustomer(email, name, address)
	if err != nil {
		return fmt.Errorf("Error creating contact in Dinero: %v", err.Error())
	}

	updatedOrder := stripe.OrderUpdateParams{}
	updatedOrder.AddMetadata(flowStatus, customerCreatedStatus)
	updatedOrder.AddMetadata("customer", contactID)
	_, err = workflow.StripeAPI.Orders.Update(o.ID, &updatedOrder)
	if err != nil {
		return fmt.Errorf("Error updating stripe order: %v", err.Error())
	}

	return nil
}

// CreateInvoice creates an invoice in the given workflow
func (workflow *ProgramWorkFlow) CreateInvoice(o stripe.Order) error {
	customerID := o.Metadata["customer"]
	amount := o.Amount
	productName := o.Items[0].Description
	invoice, err := workflow.DineroAPI.CreateInvoice(customerID, productName, amount)
	if err != nil {
		return fmt.Errorf("Dinero API - Creating invoice: %s", err.Error())
	}

	updatedOrder := stripe.OrderUpdateParams{}
	updatedOrder.AddMetadata(flowStatus, invoiceCreatedStatus)
	updatedOrder.AddMetadata("invoiceID", invoice.ID)
	updatedOrder.AddMetadata("invoiceTimestamp", invoice.Timestamp)

	_, err = workflow.StripeAPI.Orders.Update(o.ID, &updatedOrder)
	if err != nil {
		return fmt.Errorf("Stripe API - Error updating state to %v", invoiceCreatedStatus)
	}

	return nil
}

// BookInvoice knows the workflow surrounding booking an invoice in dinero and update the order accordingly
func (workflow *ProgramWorkFlow) BookInvoice(o stripe.Order) error {
	invoiceID := o.Metadata["invoiceID"]
	timestamp := o.Metadata["invoiceTimestamp"]

	invoice, err := workflow.DineroAPI.BookInvoice(invoiceID, timestamp)
	if err != nil {
		return fmt.Errorf("Dinero API - Error booking invoice: %v", err.Error())
	}

	updatedOrder := stripe.OrderUpdateParams{}
	updatedOrder.AddMetadata(flowStatus, invoiceBookedStatus)
	updatedOrder.AddMetadata("invoiceTimestamp", invoice.Timestamp)
	updatedOrder.AddMetadata("invoiceNumber", strconv.FormatInt(int64(invoice.Number), 10))

	_, err = workflow.StripeAPI.Orders.Update(o.ID, &updatedOrder)
	if err != nil {
		return fmt.Errorf("Stripe API - Error updating stripe order: %v", err.Error())
	}

	return nil
}

// CreatePayment is the workflow that creates an payment in dinero and update the stripe order to the current state
func (workflow *ProgramWorkFlow) CreatePayment(o stripe.Order) error {
	invoiceID := o.Metadata["invoiceID"]

	if err := workflow.DineroAPI.CreatePayment(invoiceID, o.Amount); err != nil {
		return fmt.Errorf("Dinero API - error creating payment: %s", err.Error())
	}

	updatedOrder := stripe.OrderUpdateParams{}
	updatedOrder.AddMetadata(flowStatus, invoicePaidStatus)

	_, err := workflow.StripeAPI.Orders.Update(o.ID, &updatedOrder)
	if err != nil {
		return fmt.Errorf("Stripe Api - Error updating state to %v", invoicePaidStatus)
	}

	return nil
}

var invoiceText = `
Hej %s

Tillykke med beslutningen om at blive en stærk og funktionel badass! Dit træningsprogram kan du download her: %s

Med træningsprogrammet er du også blevet en del af et fællesskab, hvor vi støtter, hjælper, hepper på og motiverer hinanden. Fællesskabet er udelukkende for andre som har købt programmet og træner mod samme mål. Meld dig ind med det samme lige her: http://bit.ly/2Kb9B2g

Jeg har lavet videoer af alle øvelserne så du aldrig skal være i tvivl om hvordan du gør. Dem finder du på min YouTube kanal her: http://bit.ly/2Ol7yM4

Jeg vil anbefale at du gemmer linket til YouTube-kanalen som bogmærke på din telefon, så du altid har det lige ved hånden.

Læs det hele igennem og skriv endelig til mig i Facebook gruppen hvis du har nogen spørgsmål. Ellers er det bare om at komme i gang - hvad med at starte allerede i morgen?

Ps. jeg har også vedhæftet fakturaen for dit køb på %v inkl. moms.

[link-to-pdf]

Rigtig god træning!

Kærlig hilsen
Camilla
`

// FulfillWorkflow finalizes the workflow
func (workflow *ProgramWorkFlow) FulfillWorkflow(o stripe.Order) error {
	order, err := workflow.StripeAPI.Orders.Get(o.ID, nil)
	if err != nil {
		return fmt.Errorf("Stripe API - Error getting stripe order: %s", err.Error())
	}

	if order.Status == string(stripe.OrderStatusFulfilled) {
		return nil
	}

	invoiceID := o.Metadata["invoiceID"]
	name := o.Metadata["name"]
	amount := float64(o.Amount) / 100
	downloadLink := fmt.Sprintf("https://app.camillabengtsson.dk/downloads/%v", o.ID)
	text := fmt.Sprintf(invoiceText, name, downloadLink, amount)

	if err := workflow.DineroAPI.SendInvoice(invoiceID, "Her er dit program :-)", text); err != nil {
		return fmt.Errorf("Dinero API - Error sending email: %s", err.Error())
	}

	updatedOrder := stripe.OrderUpdateParams{}
	updatedOrder.AddMetadata(flowStatus, invoiceSentStatus)
	updatedOrder.Status = stripe.String(string(stripe.OrderStatusFulfilled))

	_, err = workflow.StripeAPI.Orders.Update(o.ID, &updatedOrder)
	if err != nil {
		return fmt.Errorf("Stripe API - Error updating state to %v", invoiceSentStatus)
	}

	return nil
}

// CelebrateOrder knows how to fucking celebrate mate!
func (workflow *ProgramWorkFlow) CelebrateOrder(o stripe.Order) {
	name := o.Metadata["name"]
	slackLogging(workflow.httpClient,
		"Order "+o.ID,
		fmt.Sprintf(":gopher_dance: Well done, you just earned: %v DKK and %s will be a badass :gopher_dance:", o.Amount/100, name),
		"Completed",
		"#23D1E1")
}
