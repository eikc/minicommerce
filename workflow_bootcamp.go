package main

import (
	"fmt"

	stripe "github.com/stripe/stripe-go"
	stripeClient "github.com/stripe/stripe-go/client"
)

// BootcampWorkflow for a bootcamp
type BootcampWorkflow struct {
	DineroAPI *dineroAPI
	StripeAPI *stripeClient.API
}

var bootcampText = `
Hej %s


Ps. jeg har også vedhæftet fakturaen for dit køb på %v,- kr inkl. moms.

[link-to-pdf]

Rigtig god træning!

Kærlig hilsen
Camilla
`

// FulfillWorkflow fulfills the bootcamp workflow
func (workflow *BootcampWorkflow) FulfillWorkflow(o stripe.Order) error {
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
	text := fmt.Sprintf(bootcampText, name, amount)

	if err := workflow.DineroAPI.SendInvoice(invoiceID, "Du er tilmeldt Badass bootcamp :-)", text); err != nil {
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
