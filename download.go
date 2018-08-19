package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	stripe "github.com/stripe/stripe-go"
	"google.golang.org/appengine"
)

func download(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	orderID := params.ByName("orderid")
	ctx := appengine.NewContext(r)
	stripeAPI := getStripe(ctx)

	order, err := stripeAPI.Orders.Get(orderID, nil)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, "404 page not found")
		return
	}

	if order.Status != string(stripe.OrderStatusReturned) {
		f, err := os.Open("./strong-badass-program.pdf")
		if err != nil {
			w.WriteHeader(404)
			fmt.Fprintf(w, "404 page not found")
			return
		}
		defer f.Close()

		w.Header().Add("Content-Type", "application/pdf")
		w.Header().Add("Content-Disposition", "inline; filename=staerk-og-funktionel-badass.pdf")
		w.WriteHeader(200)
		http.ServeFile(w, r, f.Name())
		return
	}

	w.WriteHeader(404)
	fmt.Fprintf(w, "404 page not found")
}
