package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"google.golang.org/appengine/log"

	"github.com/julienschmidt/httprouter"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
)

func webhookReceiver() httprouter.Handle {

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		httpClient := getHttpClient()
		api := getClient(ctxWithTimeout)
		stripeAPI := getStripe(ctxWithTimeout)
		stripeWebhookSignature := os.Getenv("StripeWebhookSignature")

		payoutWorkflow := PayoutWorkflow{
			DineroAPI: api,
			StripeAPI: stripeAPI,
		}
		programWorkflow := &BadassWorkflow{}
		bootcampWorkflow := &BootcampWorkflow{
			StripeAPI: stripeAPI,
		}
		foodieWorkflow := &FoodieWorkflow{}
		bundleWorkflow := &BundleWorkflow{}
		onlineBootcamp := &OnlineBootcamp{
			StripeAPI: stripeAPI,
		}

		workflow := Workflow{
			Fulfillments: map[string]Fulfillment{
				"badass":         programWorkflow,
				"bootcamp":       bootcampWorkflow,
				"foodie":         foodieWorkflow,
				"bundle":         bundleWorkflow,
				"onlineBootcamp": onlineBootcamp,
			},
			DineroAPI:  api,
			StripeAPI:  stripeAPI,
			httpClient: httpClient,
		}

		var e stripe.Event
		if isDevelopmentServer() {
			decoder := json.NewDecoder(r.Body)
			decoder.Decode(&e)
		} else {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				errorHandling(w, err)
				slackLogging(httpClient, "Problems parsing body of request", err.Error(), "Error with parsing", "#CF0003")
				return
			}
			e, err = webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), stripeWebhookSignature)
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

			if err := payoutWorkflow.HandleStripePayout(p); err != nil {
				errorHandling(w, err)
				go slackLogging(httpClient, "Stripe Payout", err.Error(), "", "#CF0003")
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

			err = workflow.CreateCustomer(o)
			if err != nil {
				errorHandling(w, err)
				slackLogging(httpClient,
					fmt.Sprintf("Order %v", o.ID), err.Error(),
					fmt.Sprintf("Workflow: %v", customerCreatedStatus),
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

			workflow, err := workflow.StartFlow(o)
			if err != nil {
				log.Errorf(ctxWithTimeout, "error is the following: %v", err)
				errorHandling(w, err)
				slackLogging(httpClient,
					fmt.Sprintf("Order %v", o.ID), err.Error(),
					fmt.Sprintf("workflow: %v", workflow),
					"#CF0003")
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
