package main

import (
	"fmt"
	"os"

	"github.com/stripe/stripe-go"
)

// FoodieWorkflow handles all actions surrounding the workflow of a bundle
type FoodieWorkflow struct{}

var foodieText = `
Hej %s

Tillykke med beslutningen om at blive en badass i køkkenet!

E-bogen Camilla’s 15-minute kitchen finder du her: %s

Linket er unikt, kun til dig :-)

Jeg håber at du får glæde af mine opskrifter. Ikke mindst håber jeg også at kunne inspirere til tankegangen om, at du ikke behøver at leve udelukkende af ris og kylling for at opnå dine mål :-)

Rigtig god madlavning!

Kærlig hilsen
Camilla

______

Ps. jeg har også vedhæftet fakturaen for dit køb på %v,- kr inkl. moms.

[link-to-pdf]
`

// FulfillWorkflow finalizes the workflow
func (workflow *FoodieWorkflow) FulfillWorkflow(o stripe.Order) (string, string) {
	name := o.Metadata["name"]
	amount := float64(o.Amount) / 100
	downloadLink := fmt.Sprintf("https://app.camillabengtsson.dk/downloads/%s/%s", o.ID, "sku_DWJE6B88Ih3Wgg")
	text := fmt.Sprintf(foodieText, name, downloadLink, amount)

	return "Her er dit program :-)", text
}

func (workflow *FoodieWorkflow) GetTemplate() string {
	return os.Getenv("TEMPLATE_ONE")
}