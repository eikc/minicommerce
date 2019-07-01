package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/stripe/stripe-go"
)

type order struct {
	Name        string   `json:"name,omitempty"`
	Address     string   `json:"address,omitempty"`
	Tshirt      string   `json:"tshirt,omitempty"`
	Email       string   `json:"email,omitempty"`
	StripeToken string   `json:"stripeToken,omitempty"`
	SKU         []string `json:"sku,omitempty"`
	Newsletter  bool     `json:"newsletter,omitempty"`
}

func (o order) GetOrderType() string {
	if len(o.SKU) > 1 {
		return "bundle"
	}

	sku := o.SKU[0]

	switch sku {
	case "sku_DJx1hCHoxDAAtE":
		return "badass"
	case "sku_DWJE6B88Ih3Wgg":
		return "foodie"
	}

	return "onlineBootcamp"
}

func create() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()
		httpClient := getHttpClient()
		stripeAPI := getStripe(ctx)

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

		var itemsToOrder []*stripe.OrderItemParams
		for _, s := range o.SKU {
			item := &stripe.OrderItemParams{
				Type:   stripe.String(string(stripe.OrderItemTypeSKU)),
				Parent: stripe.String(s),
			}
			itemsToOrder = append(itemsToOrder, item)
		}

		params := &stripe.OrderParams{
			Currency: stripe.String(string(stripe.CurrencyDKK)),
			Email:    stripe.String(o.Email),
			Items:    itemsToOrder,
		}

		if len(params.Items) == 2 {
			params.Coupon = stripe.String("3Y9rWEst")
		}

		mashalledOrder, _ := json.Marshal(o)
		params.AddMetadata("name", o.Name)
		params.AddMetadata("newsletter", strconv.FormatBool(o.Newsletter))
		params.AddMetadata("rawData", string(mashalledOrder))
		params.AddMetadata("address", o.Address)
		params.AddMetadata("token", o.StripeToken)
		params.AddMetadata("email", o.Email)
		params.AddMetadata("tshirt", o.Tshirt)
		params.AddMetadata("ordertype", o.GetOrderType())

		order, err := stripeAPI.Orders.New(params)
		if err != nil {
			slackLogging(httpClient, "Could not create order", err.Error(), "Error creating order", "#CF0003")
			errorHandling(w, err)
			return
		}

		token := o.StripeToken
		op := &stripe.OrderPayParams{}
		if err := op.SetSource(token); err != nil {
			slackLogging(httpClient, "Could not prepare order payments", err.Error(), "Error paying order", "#CF0003")
			errorHandling(w, err)
			return
		} // obtained with Stripe.js

		_, err = stripeAPI.Orders.Pay(order.ID, op)
		if err != nil {
			if stripeErr, ok := err.(*stripe.Error); ok {
				var msg string

				switch stripeErr.Code {
				case stripe.ErrorCodeCardDeclined:
					msg = "Kortet er blevet afvist. Prøv igen, evt. med et andet kort."
				case stripe.ErrorCodeExpiredCard:
					msg = "Kortet er udløbet. Prøv igen med et andet kort."
				case stripe.ErrorCodeIncorrectCVC:
				case stripe.ErrorCodeInvalidCVC:
					msg = "Der er blevet indtastet forkert CVC kode. Prøv igen."
				case stripe.ErrorCodeInvalidExpiryMonth:
				case stripe.ErrorCodeInvalidExpiryYear:
					msg = "Der er blevet indtastet forkert udløbsmåned eller år. Prøv igen."
				default:
					msg = stripeErr.Msg
				}

				slackLogging(httpClient, "Could not capture payment", msg, "Error pay order", "#CF0003")
				handleStringError(w, msg)
			} else {
				slackLogging(httpClient, "Could not capture payment", err.Error(), "Error pay order", "#CF0003")
				errorHandling(w, err)
			}

			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
