package main

import (
	"fmt"
	"os"

	"github.com/stripe/stripe-go"
)

// BundleWorkflow handles all actions surrounding the workflow of a bundle
type BundleWorkflow struct{}

var bundleText = `
Hej %s

Tillykke med beslutningen om at blive en stærk og funktionel badass - både i fitnesscentret og i køkkenet!

Dit træningsprogram kan du downloade her: %s

E-bogen Camilla’s 15-minute kitchen finder du her: %s

Links’ne er unikke, kun til dig :-)

______

Med træningsprogrammet er du også blevet en del af et fællesskab, hvor vi støtter, hjælper, hepper på og motiverer hinanden. Fællesskabet er udelukkende for andre som har købt programmet og træner mod samme mål. Meld dig ind med det samme lige her: http://bit.ly/2Kb9B2g

Jeg har lavet videoer af alle øvelserne så du aldrig skal være i tvivl om hvordan du gør. Dem finder du på min YouTube kanal her: http://bit.ly/2Ol7yM4

Jeg vil anbefale at du gemmer linket til YouTube-kanalen som bogmærke på din telefon, så du altid har det lige ved hånden.

Læs det hele igennem og skriv endelig til mig i Facebook gruppen hvis du har nogen spørgsmål. Ellers er det bare om at komme i gang - hvad med at starte allerede i morgen?

Rigtig god træning!

Kærlig hilsen
Camilla

______

Ps. jeg har også vedhæftet fakturaen for dit køb på %v,- kr inkl. moms.

[link-to-pdf]
`

// FulfillWorkflow finalizes the workflow
func (workflow *BundleWorkflow) FulfillWorkflow(o stripe.Order) (string, string) {
	name := o.Metadata["name"]
	amount := float64(o.Amount) / 100
	programLink := fmt.Sprintf("https://app.camillabengtsson.dk/downloads/%s/%s", o.ID, "sku_DJx1hCHoxDAAtE")
	kitchenLink := fmt.Sprintf("https://app.camillabengtsson.dk/downloads/%s/%s", o.ID, "sku_DWJE6B88Ih3Wgg")
	text := fmt.Sprintf(bundleText, name, programLink, kitchenLink, amount)

	return "Her er dine programmer :-)", text
}

func (workflow *BundleWorkflow) GetTemplate() string {
	return os.Getenv("TEMPLATE_ONE")
}
