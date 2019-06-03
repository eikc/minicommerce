package main

import (
	"fmt"

	"github.com/stripe/stripe-go"
)

// FoodieWorkflow handles all actions surrounding the workflow of a bundle
type OnlineBootcamp struct{}

var onlineBootcampText = `
Hej %s

ONLINE BOOTCAMP!! OHH YES!

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
	text := fmt.Sprintf(onlineBootcampText, name, amount)

	return "Her er dit program :-)", text
}
