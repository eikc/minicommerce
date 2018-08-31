package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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
		payoutWorkflow := PayoutWorkflow{
			DineroAPI: api,
			StripeAPI: stripeAPI,
		}

		programWorkflow := &BadassWorkflow{
			DineroAPI: api,
			StripeAPI: stripeAPI,
		}
		bootcampWorkflow := &BootcampWorkflow{
			DineroAPI: api,
			StripeAPI: stripeAPI,
		}

		foodieWorkflow := &FoodieWorkflow{
			DineroAPI: api,
			StripeAPI: stripeAPI,
		}

		bundleWorkflow := &BundleWorkflow{
			DineroAPI: api,
			StripeAPI: stripeAPI,
		}

		workflow := Workflow{
			Fulfillments: map[string]Fulfillment{
				"badass":   programWorkflow,
				"bootcamp": bootcampWorkflow,
				"foodie":   foodieWorkflow,
				"bundle":   bundleWorkflow,
			},
			DineroAPI:  api,
			StripeAPI:  stripeAPI,
			httpClient: httpClient,
		}

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

			err = programWorkflow.CreateCustomer(o)
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
