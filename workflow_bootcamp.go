package main

import (
	"bytes"
	"fmt"

	"github.com/alecthomas/template"
	stripe "github.com/stripe/stripe-go"
	stripeClient "github.com/stripe/stripe-go/client"
)

// BootcampWorkflow for a bootcamp
type BootcampWorkflow struct {
	DineroAPI *dineroAPI
	StripeAPI *stripeClient.API
}

var bootcampText = `
Hej {{.Name}}

FEDT at du vil være med til fællestræning. Rasmus og jeg glæder os sindssygt meget til at træne med dig!

Du har tilmeldt dig fællestræning på følgende datoer:
{{ range .Items}}
{{.Date}} kl. {{.StartsAt}} - {{.Fokus}}
{{ end }} 

Fællestræningen afholdes i:
Loaded Gym
Værkstedvej 71
2500 Valby

Transport:
Der er gratis parkering lige udenfor døren.
Hvis du er med offentlig transport kan du tage S-toget til Ny Ellebjerg St. eller bus 8A til Grønttorvet (Gl. Køge Landevej).

Når du ankommer til Loaded Gym skal du henvende dig i receptionen og sige at du skal træne med Rasmus og Camilla.

Kom i god tid - vi starter til tiden!

Ps. jeg har også vedhæftet fakturaen for dit køb på {{.Amount}} inkl. moms.

[link-to-pdf]

Rigtig god træning indtil da :-)

Kærlig hilsen
Camilla
`

type emailTextForBootcamp struct {
	Name   string
	Amount float64
	Items  []chosenBootcamps
}

type chosenBootcamps struct {
	Date     string
	StartsAt string
	Fokus    string
}

// FulfillWorkflow fulfills the bootcamp workflow
func (workflow *BootcampWorkflow) FulfillWorkflow(o stripe.Order) error {
	order, err := workflow.StripeAPI.Orders.Get(o.ID, nil)
	if err != nil {
		return fmt.Errorf("Stripe API - Error getting stripe order: %s", err.Error())
	}

	if order.Status == string(stripe.OrderStatusFulfilled) {
		return nil
	}

	var bootcamps []chosenBootcamps
	for _, item := range o.Items {
		if item.Type != "sku" {
			continue
		}

		sku, _ := workflow.StripeAPI.Skus.Get(item.Parent, nil)
		date := sku.Attributes["date"]
		startsAt := sku.Attributes["StartsAt"]
		fokus := sku.Attributes["fokus"]

		b := chosenBootcamps{date, startsAt, fokus}
		bootcamps = append(bootcamps, b)
	}

	invoiceID := o.Metadata["invoiceID"]
	name := o.Metadata["name"]
	amount := float64(o.Amount) / 100
	emailText := emailTextForBootcamp{name, amount, bootcamps}

	t, err := template.New("todos").Parse(bootcampText)

	var buff bytes.Buffer
	if err := t.Execute(&buff, emailText); err != nil {
		return fmt.Errorf("Could not parse email template correctly: %s", err.Error())
	}

	if err := workflow.DineroAPI.SendInvoice(invoiceID, "Du er tilmeldt Badass bootcamp :-)", buff.String()); err != nil {
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
