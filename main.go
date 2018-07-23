package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/eikc/dinero-go/dinerotest"
	"github.com/rs/cors"

	"github.com/eikc/dinero-go"

	"github.com/julienschmidt/httprouter"
	stripe "github.com/stripe/stripe-go"
	stripeOrders "github.com/stripe/stripe-go/order"
)

const (
	flowStatus            = "flowstatus"
	customerCreatedStatus = "UserCreated"
	invoiceCreatedStatus  = "InvoiceCreated"
	invoiceBookedStatus   = "invoiceBooked"
	invoicePaidStatus     = "InvoicePaid"
	invoiceSentStatus     = "EmailSent"
)

type order struct {
	Name        string
	Address     string
	Email       string
	StripeToken string
	SKU         string
	Newsletter  bool
}

func main() {
	stripe.Key = "sk_test_XWq2CSR4oPhh80dX1QCBfs6y"
	client, secret, apiKey, orgID := dinerotest.GetClientKeysForIntegrationTesting()
	dineroClient := dinero.NewClient(client, secret)

	api := dineroAPI{
		API: dineroClient,
	}

	router := httprouter.New()
	router.GET("/", index)
	router.POST("/create", create())
	router.POST("/webhook", webhookReceiver(&api, apiKey, orgID))

	handler := cors.Default().Handler(router)

	log.Fatal(http.ListenAndServe(":8080", handler))
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func webhookReceiver(api *dineroAPI, apiKey string, orgID int) httprouter.Handle {

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		api.Authorize(apiKey, orgID)
		decoder := json.NewDecoder(r.Body)

		var e stripe.Event
		decoder.Decode(&e)

		switch e.Type {
		case "order.created":
			var o stripe.Order
			err := json.Unmarshal(e.Data.Raw, &o)
			if err != nil {
				go slackLogging(fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
				errorHandling(w, err)
				return
			}

			fmt.Println("yes an order created, paying the order")
			token := o.Metadata["token"]
			op := &stripe.OrderPayParams{}
			op.SetSource(token) // obtained with Stripe.js
			stripeOrders.Pay(o.ID, op)

			// go slackLogging(fmt.Sprintf("Order %v", o.ID), "Stripe charged succesfully", "Paid in stripe", "#2eb886")

		case "order.payment_succeeded":
			var o stripe.Order
			err := json.Unmarshal(e.Data.Raw, &o)
			if err != nil {
				go slackLogging(fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
				errorHandling(w, err)
				return
			}

			name := o.Metadata["name"]
			email := o.Metadata["email"]
			address := o.Metadata["address"]
			contactID, err := api.CreateCustomer(email, name, address)
			if err != nil {
				go slackLogging(fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
				errorHandling(w, err)
				return
			}

			updatedOrder := stripe.OrderUpdateParams{}
			updatedOrder.AddMetadata(flowStatus, customerCreatedStatus)
			updatedOrder.AddMetadata("customer", contactID)
			_, err = stripeOrders.Update(o.ID, &updatedOrder)
			if err != nil {
				go slackLogging(fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
				errorHandling(w, err)
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
				go slackLogging(fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
				errorHandling(w, err)
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
					go slackLogging(fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
					errorHandling(w, err)
					return
				}

				updatedOrder := stripe.OrderUpdateParams{}
				updatedOrder.AddMetadata(flowStatus, invoiceCreatedStatus)
				updatedOrder.AddMetadata("invoiceID", invoice.ID)
				updatedOrder.AddMetadata("invoiceTimestamp", invoice.Timestamp)

				_, err = stripeOrders.Update(o.ID, &updatedOrder)
				if err != nil {
					go slackLogging(fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
					errorHandling(w, err)
					return
				}

				// go slackLogging("Order "+o.ID, "Invoice created in draft mode", invoiceCreatedStatus, "#2eb886")

			case invoiceCreatedStatus:
				invoiceID := o.Metadata["invoiceID"]
				timestamp := o.Metadata["invoiceTimestamp"]

				invoice, err := api.BookInvoice(invoiceID, timestamp)
				if err != nil {
					go slackLogging(fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
					errorHandling(w, err)
					return
				}

				updatedOrder := stripe.OrderUpdateParams{}
				updatedOrder.AddMetadata(flowStatus, invoiceBookedStatus)
				updatedOrder.AddMetadata("invoiceTimestamp", invoice.Timestamp)
				updatedOrder.AddMetadata("invoiceNumber", strconv.FormatInt(int64(invoice.Number), 10))

				_, err = stripeOrders.Update(o.ID, &updatedOrder)
				if err != nil {
					go slackLogging(fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
					errorHandling(w, err)
					return
				}

				// go slackLogging("Order "+o.ID, fmt.Sprint("invoice booked: ", invoice.Number), invoiceBookedStatus, "#2eb886")

			case invoiceBookedStatus:
				invoiceID := o.Metadata["invoiceID"]

				if err := api.CreatePayment(invoiceID, o.Amount); err != nil {
					go slackLogging(fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
					errorHandling(w, err)
					return
				}

				updatedOrder := stripe.OrderUpdateParams{}
				updatedOrder.AddMetadata(flowStatus, invoicePaidStatus)

				_, err = stripeOrders.Update(o.ID, &updatedOrder)
				if err != nil {
					go slackLogging(fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
					errorHandling(w, err)
					return
				}

				// go slackLogging("Order "+o.ID, "Payment created for invoice", invoicePaidStatus, "#2eb886")

			case invoicePaidStatus:
				invoiceID := o.Metadata["invoiceID"]
				if err := api.SendInvoice(invoiceID); err != nil {
					go slackLogging(fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
					errorHandling(w, err)
					return
				}

				updatedOrder := stripe.OrderUpdateParams{}
				updatedOrder.AddMetadata(flowStatus, invoiceSentStatus)
				updatedOrder.Status = stripe.String(string(stripe.OrderStatusFulfilled))

				_, err = stripeOrders.Update(o.ID, &updatedOrder)
				if err != nil {
					go slackLogging(fmt.Sprintf("Order %v", o.ID), err.Error(), "Error with order", "#CF0003")
					errorHandling(w, err)
					return
				}

				// go slackLogging("Order "+o.ID, "Invoice sent to customer", invoiceSentStatus, "#2eb886")

			case invoiceSentStatus:
				go slackLogging("Order "+o.ID, fmt.Sprintf(":gopher_dance: Well done, you just earned: %v DKK :gopher_dance:", o.Amount/100), "Completed", "#23D1E1")
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

func create() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

		_, err = stripeOrders.New(params)
		if err != nil {
			errorHandling(w, err)
			return
		}

		slackLogging("New order received!", "YES! You are a BADASS!! :tada: :the_horns: :rocket:", "new", "#2eb886")
		w.WriteHeader(http.StatusOK)
	}
}

func errorHandling(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Println("error occured: ", err)
	fmt.Fprint(w, err)
	return
}

func slackLogging(title, text, status, color string) {
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

	_, err := http.Post(url, "application/json", reader)
	if err != nil {
		fmt.Println("slack error occured: ", err)
	}
}
