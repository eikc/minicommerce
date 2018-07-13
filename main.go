package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

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

	router := httprouter.New()
	router.GET("/", index)
	router.POST("/create", create())
	router.POST("/webhook", webhookReceiver)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func webhookReceiver(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

		token := o.Metadata["token"]

		fmt.Println("yes an order created, paying the order")
		op := &stripe.OrderPayParams{}
		op.SetSource(token) // obtained with Stripe.js
		stripeOrders.Pay(o.ID, op)
		fmt.Println("order paid")

	case "order.payment_succeeded":
		fmt.Println("yes order paid, sending pdf")
	case "order.payment_failed":
		fmt.Println("order payment failed, what to do!?!? :(")
	case "charge.refunded":
		fmt.Println("too bad :-(")
	}

	w.WriteHeader(http.StatusOK)
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
