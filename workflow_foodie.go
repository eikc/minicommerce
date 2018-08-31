package main

import (
	"fmt"

	"github.com/stripe/stripe-go"
	stripeClient "github.com/stripe/stripe-go/client"
)

// BundleWorkflow handles all actions surrounding the workflow of a bundle
type FoodieWorkflow struct {
	DineroAPI *dineroAPI
	StripeAPI *stripeClient.API
}

var foodieText = `
Hej %s

[link-to-pdf]

Kærlig hilsen
Camilla
`

// FulfillWorkflow finalizes the workflow
func (workflow *FoodieWorkflow) FulfillWorkflow(o stripe.Order) error {
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
	text := fmt.Sprintf(foodieText, name, downloadLink, amount)

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
