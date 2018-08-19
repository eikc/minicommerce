package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	stripe "github.com/stripe/stripe-go"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

type Bootcamp struct {
	ID        string
	Date      string
	Location  string
	StartsAt  string
	Focus     string
	SpotsLeft int64
}

func bootcamp(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := appengine.NewContext(r)
	api := getStripe(ctx)

	params := &stripe.SKUListParams{}
	params.Filters.AddFilter("limit", "", "100")
	params.Product = stripe.String("prod_DPZ6WIQmyderfW")
	params.InStock = stripe.Bool(true)

	var bootcamps []Bootcamp

	i := api.Skus.List(params)
	for i.Next() {
		x := i.SKU()
		date := x.Attributes["date"]
		location := x.Attributes["location"]
		startsAt := x.Attributes["StartsAt"]
		focus := x.Attributes["fokus"]
		res, _ := time.Parse("02-01-2006", date)
		if res.After(time.Now()) {
			b := Bootcamp{x.ID, date, location, startsAt, focus, x.Inventory.Quantity}
			bootcamps = append(bootcamps, b)
		}
	}

	b, _ := json.Marshal(&bootcamps)

	w.Header().Add("Content-Type", "Application/json")
	fmt.Fprint(w, string(b))
}

type bootcampOrder struct {
	Name        string   `json:"name,omitempty"`
	Email       string   `json:"email,omitempty"`
	SKU         []string `json:"sku,omitempty"`
	StripeToken string   `json:"stripeToken,omitempty"`
}

func buyBootcamp(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := appengine.NewContext(r)
	httpClient := urlfetch.Client(ctx)
	stripeAPI := getStripe(ctx)

	var o bootcampOrder
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&o)
	if err != nil {
		errorHandling(w, err)
		return
	}

	var items []*stripe.OrderItemParams

	for _, ol := range o.SKU {
		item := &stripe.OrderItemParams{
			Type:   stripe.String(string(stripe.OrderItemTypeSKU)),
			Parent: stripe.String(ol),
		}

		items = append(items, item)
	}

	params := &stripe.OrderParams{
		Currency: stripe.String(string(stripe.CurrencyDKK)),
		Email:    stripe.String(o.Email),
		Items:    items,
	}

	params.AddMetadata("name", o.Name)
	params.AddMetadata("ordertype", "bootcamp")
	params.AddMetadata("token", o.StripeToken)
	params.AddMetadata("email", o.Email)

	_, err = stripeAPI.Orders.New(params)
	if err != nil {
		slackLogging(httpClient, "Could not create bootcamp order", err.Error(), "Error creating bootcamp order", "#CF0003")
		errorHandling(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
