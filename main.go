package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/eikc/dinero-go/contacts"
	"github.com/eikc/dinero-go/invoices"

	"github.com/eikc/dinero-go/dinerotest"

	"github.com/eikc/dinero-go"

	"github.com/julienschmidt/httprouter"
	stripe "github.com/stripe/stripe-go"
	stripeOrders "github.com/stripe/stripe-go/order"
)

type order struct {
	Name        string
	Address     string
	Email       string
	StripeToken string
	SKU         string
	Newsletter  bool
}

func main() {
	stripe.Key = "sk_test_XWq2CSR4oPhh80dX1QCBfs6y"
	client, secret, apiKey, orgID := dinerotest.GetClientKeysForIntegrationTesting()
	dineroClient := dinero.NewClient(client, secret)
	dineroClient.Authorize(apiKey, orgID)

	router := httprouter.New()
	router.GET("/", index)
	router.POST("/create", create())
	router.POST("/webhook", webhookReceiver(dineroClient))

	log.Fatal(http.ListenAndServe(":8080", router))
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func webhookReceiver(api dinero.API) httprouter.Handle {

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		decoder := json.NewDecoder(r.Body)

		var e stripe.Event
		decoder.Decode(&e)

		fmt.Println("event is: ", e)

		switch e.Type {
		case "order.created":
			var o stripe.Order
			err := json.Unmarshal(e.Data.Raw, &o)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			fmt.Println("creating dinero invoice..")
			cParams := contacts.ContactParams{
				IsPerson:             true,
				Name:                 o.Metadata["name"],
				Email:                o.Email,
				Street:               o.Metadata["address"],
				CountryKey:           "DK",
				PaymentConditionType: dinero.NettoCash,
			}

			c, err := contacts.Add(api, cParams)
			if err != nil {
				fmt.Println("what should we do? err: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			invoiceParams := invoices.CreateInvoice{
				ContactID:        c.ID,
				ShowLinesInclVat: true,
				Currency:         "DKK",
				Language:         "da-DK",
				Description:      "Awesome badass program",
				Date:             dinero.DateNow(),
				ProductLines: []invoices.InvoiceLine{
					invoices.InvoiceLine{
						BaseAmountValue: float64((o.Amount / 100)),
						Quantity:        1,
						AccountNumber:   1000,
						Description:     "Fitness program",
						LineType:        "Product",
						Unit:            "parts",
					},
				},
			}

			invoice, err := invoices.Save(api, invoiceParams)
			if err != nil {
				fmt.Println("what should we do? err: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			fmt.Println("yes an order created, paying the order")
			token := o.Metadata["token"]
			op := &stripe.OrderPayParams{}
			op.SetSource(token) // obtained with Stripe.js
			op.AddMetadata("invoiceID", invoice.ID)
			op.AddMetadata("invoiceTimestamp", invoice.Timestamp)
			stripeOrders.Pay(o.ID, op)
			fmt.Println("order paid, invoice created")

		case "order.payment_succeeded":
			var o stripe.Order
			err := json.Unmarshal(e.Data.Raw, &o)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			invoiceID := o.Metadata["invoiceID"]
			invoiceTimestamp := o.Metadata["invoiceTimestamp"]

			timestamp, err := invoices.Book(api, invoiceID, invoiceTimestamp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Println("booking invoice failed: ", err)
				return
			}

			payment := invoices.CreatePaymentParams{
				Amount:               float64(o.Amount / 100),
				DepositAccountNumber: 55000,
				Description:          "Paid with stripe",
				PaymentDate:          dinero.DateNow(),
				Timestamp:            timestamp.Timestamp,
			}

			timestamp, err = invoices.CreatePayment(api, invoiceID, payment)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Println("paying invoice failed: ", err)
				return
			}

			fmt.Println("sending invoice...")
			_, err = invoices.SendEmail(api, invoiceID, invoices.SendInvoice{})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Println("sending invoice failed ", err)
				return
			}

			fmt.Println("yes order paid, sending pdf")
		case "order.payment_failed":
			fmt.Println("order payment failed, what to do!?!? :(")
		case "charge.refunded":
			fmt.Println("too bad :-(")
		}

		w.WriteHeader(http.StatusOK)
	}
}

func create() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var o order
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&o)
		if err != nil {
			fmt.Fprint(w, err)
		}

		params := &stripe.OrderParams{
			Currency: stripe.String(string(stripe.CurrencyDKK)),
			Email:    stripe.String(o.Email),
			Items: []*stripe.OrderItemParams{
				&stripe.OrderItemParams{
					Type:   stripe.String(string(stripe.OrderItemTypeSKU)),
					Parent: stripe.String(o.SKU),
				},
			},
		}
		mashalledOrder, _ := json.Marshal(o)
		params.AddMetadata("name", o.Name)
		params.AddMetadata("newsletter", strconv.FormatBool(o.Newsletter))
		params.AddMetadata("rawData", string(mashalledOrder))
		params.AddMetadata("address", o.Address)
		params.AddMetadata("token", o.StripeToken)

		_, err = stripeOrders.New(params)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
