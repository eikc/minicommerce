package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	stripe "github.com/stripe/stripe-go"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

func bootcamp(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := appengine.NewContext(r)
	api := getStripe(ctx)

	params := &stripe.SKUListParams{}
	params.Filters.AddFilter("limit", "", "100")
	params.Product = stripe.String("prod_DPZ6WIQmyderfW")
	params.InStock = stripe.Bool(true)

	var test []Bootcamp

	i := api.Skus.List(params)
	for i.Next() {
		x := i.SKU()
		log.Infof(ctx, "it works: %v", x)
		date := x.Attributes["date"]
		location := x.Attributes["location"]
		startsAt := x.Attributes["StartsAt"]
		res, _ := time.Parse("2006-01-02", date)
		if res.After(time.Now()) {
			b := Bootcamp{x.ID, date, location, startsAt, x.Inventory.Quantity}
			test = append(test, b)
		}
	}

	b, _ := json.Marshal(&test)

	fmt.Fprint(w, string(b))
}
