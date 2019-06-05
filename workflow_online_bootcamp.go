package main

import (
	"fmt"

	"github.com/stripe/stripe-go"
	stripeClient "github.com/stripe/stripe-go/client"
)

// OnlineBootcamp handles all actions surrounding the workflow of a bundle
type OnlineBootcamp struct {
	StripeAPI *stripeClient.API
}

var onlineBootcampText = `
Hej %s

Tak for din tilmelding til Boss Babes Online Bootcamp. Vi glæder os virkelig meget til 4 uger sammen med dig.

Du kan tilmelde dig Facebook gruppen for månedens bootcamp her: %s

Det er super vigtigt at du tilmelder dig gruppen med det samme, da alt du behøver at vide bliver delt inde i gruppen.

Dit træningsprogram og kostplan bliver uploadet i gruppen i weekenden før bootcampen starter.

Rigtig god træning indtil da 💪

Kh
Rasmus og Camilla

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
