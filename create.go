package main

import (
	"encoding/json"
	"io/ioutil"
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

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			slackLogging(httpClient, "Error reading all from request body", err.Error(), "Incomplete order", "#CF0003")
			errorHandling(w, err)
			return
		}

		defer r.Body.Close()

		var o order
		err = json.Unmarshal(body, &o)
		if err != nil {
			slackLogging(httpClient, "Error decoding response from program buy", string(body), "Incomplete order", "#CF0003")
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
		params.AddMetadata("ordertype", "badass")

		_, err = stripeAPI.Orders.New(params)
		if err != nil {
			slackLogging(httpClient, "Could not create order", err.Error(), "Error creating order", "#CF0003")
			errorHandling(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
