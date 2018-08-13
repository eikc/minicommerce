package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

func webhookReceiver() httprouter.Handle {

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		c := appengine.NewContext(r)
		s := getSettings(c)

		httpClient := urlfetch.Client(c)

		api := getClient(c)
		stripeAPI := getStripe(c)

		var e stripe.Event
		if appengine.IsDevAppServer() {
			decoder := json.NewDecoder(r.Body)
			decoder.Decode(&e)
		} else {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				errorHandling(w, err)
				slackLogging(httpClient, "Problems parsing body of request", err.Error(), "Error with parsing", "#CF0003")
				return
			}
			e, err = webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), s.StripeWebhookSignature)
			if err != nil {
				errorHandling(w, err)
				slackLogging(httpClient, "Problems constructing stripe event", err.Error(), "Event Error", "#CF0003")
				return
			}
		}

		switch e.Type {
		case "payout.paid":
			var p stripe.Payout
			if err := json.Unmarshal(e.Data.Raw, &p); err != nil {
				errorHandling(w, err)
				go slackLogging(httpClient, "Problems constructing Payout event", err.Error(), "Event error - Payout", "#CF0003")
				return
			}
			params := &stripe.BalanceTransactionListParams{}
			params.Payout = stripe.String(p.ID)

			i := stripeAPI.Balance.List(params)

			var feeAmount int64
			for i.Next() {
				bt := i.BalanceTransaction()
				feeAmount += bt.Fee
			}

			amount := float64(p.Amount) / 100
			fee := float64(feeAmount) / 100

			err := api.AddStripePayout(p.ID, amount, fee)

			if err != nil {
				errorHandling(w, err)
				go slackLogging(httpClient, "Stripe Payout", err.Error(), "Failed to update ledger in Dinero", "#CF0003")
				return
			}

			slackLogging(httpClient,
				"payout paid: "+p.ID,
				fmt.Sprintf("money money money... it's so funny, your bank is filled with: %v", p.Amount/100),
				"Completed",
				"#23D1E1")

		case "order.created":
			var o stripe.Order
			err := json.Unmarshal(e.Data.Raw, &o)
			if err != nil {
				errorHandling(w, err)
				go slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order under created flow", "#CF0003")
				return
			}

			fmt.Println("yes an order created, paying the order")
			token := o.Metadata["token"]
			op := &stripe.OrderPayParams{}
			op.SetSource(token) // obtained with Stripe.js
			_, err = stripeAPI.Orders.Pay(o.ID, op)
			if err != nil {
				slackLogging(httpClient, "Stripe charge failed", fmt.Sprint("stripe charge failed: ", err.Error()), "Stripe charge failed", "#CF0003")
			}

		case "order.payment_succeeded":
			var o stripe.Order
			err := json.Unmarshal(e.Data.Raw, &o)
			if err != nil {
				errorHandling(w, err)
				go slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Error unmarshalling order! PANIC!", "#CF0003")
				return
			}

			name := o.Metadata["name"]
			email := o.Metadata["email"]
			address := o.Metadata["address"]
			contactID, err := api.CreateCustomer(email, name, address)
			if err != nil {
				errorHandling(w, err)
				go slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Error creating contact in Dinero", "#CF0003")
				return
			}

			updatedOrder := stripe.OrderUpdateParams{}
			updatedOrder.AddMetadata(flowStatus, customerCreatedStatus)
			updatedOrder.AddMetadata("customer", contactID)
			_, err = stripeAPI.Orders.Update(o.ID, &updatedOrder)
			if err != nil {
				errorHandling(w, err)
				slackLogging(httpClient,
					fmt.Sprintf("Order %v", o.ID), err.Error(),
					fmt.Sprintf("Stripe API - Error updating state to %v", customerCreatedStatus),
					"#CF0003")
				return
			}

		case "order.payment_failed":
			fmt.Println("order payment failed, what to do!?!? :(")
		case "charge.refunded":
			fmt.Println("too bad :-(")

		case "order.updated":
			var o stripe.Order
			err := json.Unmarshal(e.Data.Raw, &o)
			if err != nil {
				errorHandling(w, err)
				slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "error with unmarshalling order, PANIC!", "#CF0003")
				return
			}

			flow := o.Metadata[flowStatus]

			switch flow {
			case customerCreatedStatus:
				customerID := o.Metadata["customer"]
				amount := o.Amount
				productName := o.Items[0].Description
				invoice, err := api.CreateInvoice(customerID, productName, amount)
				if err != nil {
					errorHandling(w, err)
					slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Error talking to Dinero - Creating invoice", "#CF0003")
					return
				}

				updatedOrder := stripe.OrderUpdateParams{}
				updatedOrder.AddMetadata(flowStatus, invoiceCreatedStatus)
				updatedOrder.AddMetadata("invoiceID", invoice.ID)
				updatedOrder.AddMetadata("invoiceTimestamp", invoice.Timestamp)

				_, err = stripeAPI.Orders.Update(o.ID, &updatedOrder)
				if err != nil {
					errorHandling(w, err)
					slackLogging(httpClient,
						fmt.Sprintf("Order %v", o.ID), err.Error(),
						fmt.Sprintf("Stripe API - Error updating state to %v", invoiceCreatedStatus),
						"#CF0003")
					return
				}

			case invoiceCreatedStatus:
				invoiceID := o.Metadata["invoiceID"]
				timestamp := o.Metadata["invoiceTimestamp"]

				invoice, err := api.BookInvoice(invoiceID, timestamp)
				if err != nil {
					errorHandling(w, err)
					slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Dinero api - Error booking invoice", "#CF0003")
					return
				}

				updatedOrder := stripe.OrderUpdateParams{}
				updatedOrder.AddMetadata(flowStatus, invoiceBookedStatus)
				updatedOrder.AddMetadata("invoiceTimestamp", invoice.Timestamp)
				updatedOrder.AddMetadata("invoiceNumber", strconv.FormatInt(int64(invoice.Number), 10))

				_, err = stripeAPI.Orders.Update(o.ID, &updatedOrder)
				if err != nil {
					errorHandling(w, err)
					slackLogging(httpClient,
						fmt.Sprintf("Order %v", o.ID), err.Error(),
						fmt.Sprintf("Stripe API - Error updating state to %v", invoiceBookedStatus),
						"#CF0003")
					return
				}

			case invoiceBookedStatus:
				invoiceID := o.Metadata["invoiceID"]

				if err := api.CreatePayment(invoiceID, o.Amount); err != nil {
					errorHandling(w, err)
					slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Dinero API - Error creating payment", "#CF0003")
					return
				}

				updatedOrder := stripe.OrderUpdateParams{}
				updatedOrder.AddMetadata(flowStatus, invoicePaidStatus)

				_, err = stripeAPI.Orders.Update(o.ID, &updatedOrder)
				if err != nil {
					errorHandling(w, err)
					slackLogging(httpClient,
						fmt.Sprintf("Order %v", o.ID), err.Error(),
						fmt.Sprintf("Stripe API - Error updating state to %v", invoicePaidStatus),
						"#CF0003")
					return
				}

			case invoicePaidStatus:
				order, err := stripeAPI.Orders.Get(o.ID, nil)
				if err != nil {
					errorHandling(w, err)
					slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Stripe API - error getting stripe order", "#CF0003")
					return
				}

				if order.Status == string(stripe.OrderStatusFulfilled) {
					break
				}

				invoiceID := o.Metadata["invoiceID"]
				if err := api.SendInvoice(invoiceID); err != nil {
					errorHandling(w, err)
					slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Dinero API - Error sending email", "#CF0003")
					return
				}

				updatedOrder := stripe.OrderUpdateParams{}
				updatedOrder.AddMetadata(flowStatus, invoiceSentStatus)
				updatedOrder.Status = stripe.String(string(stripe.OrderStatusFulfilled))

				_, err = stripeAPI.Orders.Update(o.ID, &updatedOrder)
				if err != nil {
					errorHandling(w, err)
					slackLogging(httpClient,
						fmt.Sprintf("Order %v", o.ID), err.Error(),
						fmt.Sprintf("Stripe API - Error updating state to %v", invoiceSentStatus),
						"#CF0003")
					return
				}

			case invoiceSentStatus:
				name := o.Metadata["name"]
				slackLogging(httpClient,
					"Order "+o.ID,
					fmt.Sprintf(":gopher_dance: Well done, you just earned: %v DKK and %s will be a badass :gopher_dance:", o.Amount/100, name),
					"Completed",
					"#23D1E1")
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
