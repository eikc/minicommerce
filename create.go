package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	stripe "github.com/stripe/stripe-go"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

type order struct {
	Name        string
	Address     string
	Email       string
	StripeToken string
	SKU         string
	Newsletter  bool
}

func create() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		c := appengine.NewContext(r)
		httpClient := urlfetch.Client(c)
		stripeAPI := getStripe(c)

		var o order
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&o)
		if err != nil {
			errorHandling(w, err)
			return
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
		params.AddMetadata("email", o.Email)

		_, err = stripeAPI.Orders.New(params)
		if err != nil {
			slackLogging(httpClient, "Could not create order", err.Error(), "Error creating order", "#CF0003")
			errorHandling(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
