package main

import (
	"fmt"

	"github.com/stripe/stripe-go"
	stripeClient "github.com/stripe/stripe-go/client"
)

// ProgramWorkFlow handles all actions surrounding the workflow of a program
type ProgramWorkFlow struct {
	DineroAPI *dineroAPI
	StripeAPI *stripeClient.API
}

var badassText = `
Hej %s

Tillykke med beslutningen om at blive en stærk og funktionel badass! Dit træningsprogram kan du downloade her: %s

Med træningsprogrammet er du også blevet en del af et fællesskab, hvor vi støtter, hjælper, hepper på og motiverer hinanden. Fællesskabet er udelukkende for andre som har købt programmet og træner mod samme mål. Meld dig ind med det samme lige her: http://bit.ly/2Kb9B2g

Jeg har lavet videoer af alle øvelserne så du aldrig skal være i tvivl om hvordan du gør. Dem finder du på min YouTube kanal her: http://bit.ly/2Ol7yM4

Jeg vil anbefale at du gemmer linket til YouTube-kanalen som bogmærke på din telefon, så du altid har det lige ved hånden.

Læs det hele igennem og skriv endelig til mig i Facebook gruppen hvis du har nogen spørgsmål. Ellers er det bare om at komme i gang - hvad med at starte allerede i morgen?

Ps. jeg har også vedhæftet fakturaen for dit køb på %v,- kr inkl. moms.

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
	text := fmt.Sprintf(badassText, name, downloadLink, amount)

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
