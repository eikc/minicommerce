package main

import (
	"fmt"

	"github.com/stripe/stripe-go"
	stripeClient "github.com/stripe/stripe-go/client"
)

// FoodieWorkflow handles all actions surrounding the workflow of a bundle
type OnlineBootcamp struct {
	StripeAPI *stripeClient.API
}

var onlineBootcampText = `
Hej %s

ONLINE BOOTCAMP!! OHH YES!

Facebook gruppen findes her: %s

Kærlig hilsen
Rasmus & Camilla

______

Ps. jeg har også vedhæftet fakturaen for dit køb på %v,- kr inkl. moms.

[link-to-pdf]
`

// FulfillWorkflow finalizes the workflow
func (workflow *OnlineBootcamp) FulfillWorkflow(o stripe.Order) (string, string) {
	name := o.Metadata["name"]
	amount := float64(o.Amount) / 100
	var facebook string

	for _, item := range o.Items {
		if item.Type != "sku" {
			continue
		}

		sku, _ := workflow.StripeAPI.Skus.Get(item.Parent, nil)
		facebook = sku.Attributes["facebook"]
	}

	text := fmt.Sprintf(onlineBootcampText, name, facebook, amount)

	return "Her er dit program :-)", text
}
