package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/stripe/stripe-go"
)

func download(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	orderID := params.ByName("orderid")
	ctx := r.Context()
	stripeAPI := getStripe(ctx)

	order, err := stripeAPI.Orders.Get(orderID, nil)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, "404 page not found")
		return
	}

	if order.Status != string(stripe.OrderStatusReturned) {
		var found bool
		for _, item := range order.Items {
			if item.Parent.SKU.ID == "sku_DJx1hCHoxDAAtE" {
				found = true
			}
		}

		if found {
			f, err := os.Open("./sku_DJx1hCHoxDAAtE.pdf")
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
	}

	w.WriteHeader(404)
	fmt.Fprintf(w, "404 page not found")
}

func downloadV2(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	orderID := params.ByName("orderid")
	skuID := params.ByName("sku")

	ctx := r.Context()

	stripeAPI := getStripe(ctx)

	order, err := stripeAPI.Orders.Get(orderID, nil)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, "404 page not found")
		return
	}

	if order.Status == string(stripe.OrderStatusReturned) {
		w.WriteHeader(404)

		fmt.Fprintf(w, "404 page not found")
		return
	}

	var found bool
	description := "program"
	for _, item := range order.Items {
		if item.Type != stripe.OrderItemTypeSKU {
			continue
		}

		if item.Parent.ID == skuID {
			found = true
			description = item.Description
			break
		}
	}

	if found {
		f, err := os.Open(fmt.Sprintf("./%s.pdf", skuID))
		if err != nil {
			w.WriteHeader(404)
			fmt.Fprintf(w, "404 page not found")
			return
		}
		defer f.Close()

		w.Header().Add("Content-Type", "application/pdf")
		w.Header().Add("Content-Disposition", fmt.Sprintf("inline; filename=%s.pdf", description))
		http.ServeFile(w, r, f.Name())
		return
	}

	w.WriteHeader(404)
	fmt.Fprintf(w, "404 page not found")
}
