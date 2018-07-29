package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/rs/cors"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"github.com/eikc/dinero-go"

	"github.com/julienschmidt/httprouter"
	stripe "github.com/stripe/stripe-go"
	stripeClient "github.com/stripe/stripe-go/client"
	"github.com/stripe/stripe-go/webhook"
)

const (
	flowStatus            = "flowstatus"
	customerCreatedStatus = "UserCreated"
	invoiceCreatedStatus  = "InvoiceCreated"
	invoiceBookedStatus   = "invoiceBooked"
	invoicePaidStatus     = "InvoicePaid"
	invoiceSentStatus     = "EmailSent"
)

type Settings struct {
	ClientKey              string
	ClientSecret           string
	APIKey                 string
	OrganizationID         int
	StripeKey              string
	StripeWebhookSignature string
	InstagramToken         string
}

type order struct {
	Name        string
	Address     string
	Email       string
	StripeToken string
	SKU         string
	Newsletter  bool
}

func main() {
	router := httprouter.New()
	router.GET("/", index)
	router.POST("/create", create())
	router.POST("/webhook", webhookReceiver())
	router.GET("/instagram", instagram)

	handler := cors.Default().Handler(router)

	http.Handle("/", handler)
	appengine.Main()
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "test")
}

func instagram(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)
	settings := getSettings(ctx)

	url := fmt.Sprint("https://api.instagram.com/v1/users/self/media/recent/?access_token=", settings.InstagramToken)

	resp, err := client.Get(url)
	if err != nil {
		log.Errorf(ctx, "instagram error occured! %v", err)
		errorHandling(w, err)
		return
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf(ctx, "instagram error occured! %v", err)
		errorHandling(w, err)
		return
	}

	var userResp UserResponse
	if err := json.Unmarshal(b, &userResp); err != nil {
		log.Errorf(ctx, "instagram error occured! %v", err)
		errorHandling(w, err)
		return
	}

	userResp.User = userResp.User[:6]

	json, err := json.Marshal(userResp.User)
	if err != nil {
		log.Errorf(ctx, "instagram error occured! %v", err)
		errorHandling(w, err)
		return
	}

	w.Header().Add("Content-Type", "Application/json")
	fmt.Fprint(w, string(json))
}

func getClient(ctx context.Context) *dineroAPI {
	key := datastore.NewKey(ctx, "settings", "settings", 0, nil)

	var s Settings
	datastore.Get(ctx, key, &s)

	httpClient := urlfetch.Client(ctx)
	dineroClient := dinero.NewClient(s.ClientKey, s.ClientSecret, httpClient)

	log.Debugf(ctx, "%v", s)
	api := dineroAPI{
		API: dineroClient,
	}

	if err := api.Authorize(s.APIKey, s.OrganizationID); err != nil {
		log.Criticalf(ctx, "Can't authorize with dinero, settings: %v - err: %v", s, err)
		panic(err)
	}

	return &api
}

func getStripe(ctx context.Context) *stripeClient.API {
	httpClient := urlfetch.Client(ctx)
	s := getSettings(ctx)

	sc := stripeClient.New(s.StripeKey, stripe.NewBackends(httpClient))

	return sc
}

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
				go slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
				return
			}

			name := o.Metadata["name"]
			email := o.Metadata["email"]
			address := o.Metadata["address"]
			log.Debugf(c, "creating contact...")
			contactID, err := api.CreateCustomer(email, name, address)
			if err != nil {
				errorHandling(w, err)
				go slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
				return
			}
			log.Debugf(c, "updating order...")

			updatedOrder := stripe.OrderUpdateParams{}
			updatedOrder.AddMetadata(flowStatus, customerCreatedStatus)
			updatedOrder.AddMetadata("customer", contactID)
			_, err = stripeAPI.Orders.Update(o.ID, &updatedOrder)
			if err != nil {
				errorHandling(w, err)
				slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
				return
			}

			// go slackLogging("Order "+o.ID, "Contact created in dinero", customerCreatedStatus, "#2eb886")

		case "order.payment_failed":
			fmt.Println("order payment failed, what to do!?!? :(")
		case "charge.refunded":
			fmt.Println("too bad :-(")

		case "order.updated":
			var o stripe.Order
			err := json.Unmarshal(e.Data.Raw, &o)
			if err != nil {
				errorHandling(w, err)
				slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
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
					slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
					return
				}

				updatedOrder := stripe.OrderUpdateParams{}
				updatedOrder.AddMetadata(flowStatus, invoiceCreatedStatus)
				updatedOrder.AddMetadata("invoiceID", invoice.ID)
				updatedOrder.AddMetadata("invoiceTimestamp", invoice.Timestamp)

				_, err = stripeAPI.Orders.Update(o.ID, &updatedOrder)
				if err != nil {
					errorHandling(w, err)
					slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
					return
				}

				// go slackLogging("Order "+o.ID, "Invoice created in draft mode", invoiceCreatedStatus, "#2eb886")

			case invoiceCreatedStatus:
				invoiceID := o.Metadata["invoiceID"]
				timestamp := o.Metadata["invoiceTimestamp"]

				invoice, err := api.BookInvoice(invoiceID, timestamp)
				if err != nil {
					errorHandling(w, err)
					slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
					return
				}

				updatedOrder := stripe.OrderUpdateParams{}
				updatedOrder.AddMetadata(flowStatus, invoiceBookedStatus)
				updatedOrder.AddMetadata("invoiceTimestamp", invoice.Timestamp)
				updatedOrder.AddMetadata("invoiceNumber", strconv.FormatInt(int64(invoice.Number), 10))

				_, err = stripeAPI.Orders.Update(o.ID, &updatedOrder)
				if err != nil {
					errorHandling(w, err)
					slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
					return
				}

				// go slackLogging("Order "+o.ID, fmt.Sprint("invoice booked: ", invoice.Number), invoiceBookedStatus, "#2eb886")

			case invoiceBookedStatus:
				invoiceID := o.Metadata["invoiceID"]

				if err := api.CreatePayment(invoiceID, o.Amount); err != nil {
					errorHandling(w, err)
					slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
					return
				}

				updatedOrder := stripe.OrderUpdateParams{}
				updatedOrder.AddMetadata(flowStatus, invoicePaidStatus)

				_, err = stripeAPI.Orders.Update(o.ID, &updatedOrder)
				if err != nil {
					errorHandling(w, err)
					slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
					return
				}

				// go slackLogging("Order "+o.ID, "Payment created for invoice", invoicePaidStatus, "#2eb886")

			case invoicePaidStatus:
				invoiceID := o.Metadata["invoiceID"]
				if err := api.SendInvoice(invoiceID); err != nil {
					errorHandling(w, err)
					slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
					return
				}

				updatedOrder := stripe.OrderUpdateParams{}
				updatedOrder.AddMetadata(flowStatus, invoiceSentStatus)
				updatedOrder.Status = stripe.String(string(stripe.OrderStatusFulfilled))

				_, err = stripeAPI.Orders.Update(o.ID, &updatedOrder)
				if err != nil {
					errorHandling(w, err)
					slackLogging(httpClient, fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
					return
				}

				// go slackLogging("Order "+o.ID, "Invoice sent to customer", invoiceSentStatus, "#2eb886")

			case invoiceSentStatus:
				slackLogging(httpClient, "Order "+o.ID, fmt.Sprintf(":gopher_dance: Well done, you just earned: %v DKK :gopher_dance:", o.Amount/100), "Completed", "#23D1E1")
			}
		}

		w.WriteHeader(http.StatusOK)
	}
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

func errorHandling(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Sprintln("error occured: ", err)
	fmt.Fprint(w, err)
	return
}

func slackLogging(httpClient *http.Client, title, text, status, color string) {
	url := "https://hooks.slack.com/services/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"
	attachment1 := Attachment{}
	attachment1.addField(Field{Title: "Title", Value: title})
	attachment1.addField(Field{Title: "Status", Value: status})
	attachment1.addField(Field{Title: "Extra info", Value: text})
	attachment1.AuthorIcon = stripe.String(":gopher_dance:")
	attachment1.Color = stripe.String(color)

	payload := Payload{
		Username:    "robot",
		Channel:     "#logging",
		IconEmoji:   ":gopher_dance:",
		Attachments: []Attachment{attachment1},
	}

	json, _ := json.Marshal(payload)
	reader := bytes.NewReader(json)

	_, err := httpClient.Post(url, "application/json", reader)
	if err != nil {
		fmt.Println("slack error occured: ", err)
	}
}
