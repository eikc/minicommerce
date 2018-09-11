package main

import (
	"fmt"
	"net/http"
	"strconv"

	stripe "github.com/stripe/stripe-go"
	stripeClient "github.com/stripe/stripe-go/client"
)

// Fulfillment is how the order is fulfilled
type Fulfillment interface {
	FulfillWorkflow(o stripe.Order) (string, string)
}

// Workflow is how stripe is used to create a workflow surrounding orders of different types
type Workflow struct {
	Fulfillments map[string]Fulfillment
	DineroAPI    *dineroAPI
	StripeAPI    *stripeClient.API
	httpClient   *http.Client
}

// StartFlow Controls the workflow of the order
func (workflow *Workflow) StartFlow(o stripe.Order) (string, error) {
	flow := o.Metadata[flowStatus]
	var err error

	oP, err := workflow.StripeAPI.Orders.Get(o.ID, nil)
	if err != nil {
		return "", err
	}

	if oP.Status == string(stripe.OrderStatusCanceled) || oP.Status == string(stripe.OrderStatusReturned) {
		return "", nil
	}

	switch flow {
	case customerCreatedStatus:
		err = workflow.CreateInvoice(o)
	case invoiceCreatedStatus:
		err = workflow.BookInvoice(o)
	case invoiceBookedStatus:
		err = workflow.CreatePayment(o)
	case invoicePaidStatus:
		err = workflow.SendInvoice(o)
	case invoiceSentStatus:
		workflow.CelebrateOrder(o)
	}

	return flow, err
}

// CreateCustomer Creates a customer in the given workflow
func (workflow *Workflow) CreateCustomer(o stripe.Order) error {
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
func (workflow *Workflow) CreateInvoice(o stripe.Order) error {
	customerID := o.Metadata["customer"]
	var lines []InvoiceLine

	for _, l := range o.Items {
		if l.Type == "sku" {
			line := InvoiceLine{
				Amount:      l.Amount,
				Description: l.Description,
			}

			lines = append(lines, line)
		}

		if l.Type == "discount" {
			line := InvoiceLine{
				Amount:      l.Amount,
				Description: "Rabat",
			}

			lines = append(lines, line)
		}
	}

	invoice, err := workflow.DineroAPI.CreateInvoice(customerID, lines)
	if err != nil {
		return fmt.Errorf("Dinero API - Creating invoice: %s - Lines are: %v", err.Error(), lines)
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
func (workflow *Workflow) BookInvoice(o stripe.Order) error {
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
func (workflow *Workflow) CreatePayment(o stripe.Order) error {
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

func (workflow *Workflow) SendInvoice(o stripe.Order) error {
	order, err := workflow.StripeAPI.Orders.Get(o.ID, nil)
	if err != nil {
		return fmt.Errorf("Stripe API - Error getting stripe order: %s", err.Error())
	}

	if order.Status == string(stripe.OrderStatusFulfilled) {
		return nil
	}

	invoiceID := o.Metadata["invoiceID"]
	ordertype := o.Metadata["ordertype"]
	handler := workflow.Fulfillments[ordertype]
	if handler == nil {
		return fmt.Errorf("No handler for ordertype")
	}

	title, text := handler.FulfillWorkflow(o)
	if err := workflow.DineroAPI.SendInvoice(invoiceID, title, text); err != nil {
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
func (workflow *Workflow) CelebrateOrder(o stripe.Order) {
	name := o.Metadata["name"]
	ordertype := o.Metadata["ordertype"]
	slackLogging(workflow.httpClient,
		fmt.Sprintf("Order %s - type %s", o.ID, ordertype),
		fmt.Sprintf("Well done, you just earned: %v DKK and %s will be a badass", o.Amount/100, name),
		"Completed",
		"#23D1E1")
}
