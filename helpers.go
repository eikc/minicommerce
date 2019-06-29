package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/eikc/dinero-go"
	"github.com/stripe/stripe-go"
	stripeClient "github.com/stripe/stripe-go/client"
)

func getClient(ctx context.Context) *dineroAPI {
	apiKey := os.Getenv("CLIENTAPIKEY")
	clientKey := os.Getenv("CLIENTKEY")
	clientSecret := os.Getenv("CLIENTSECRET")
	organizationID, _ := strconv.ParseInt(os.Getenv("CLIENTORGANIZATIONID"), 10, 64)

	httpClient := getHttpClient()
	dineroClient := dinero.NewClient(clientKey, clientSecret, httpClient)

	api := dineroAPI{
		API: dineroClient,
	}

	if err := api.Authorize(apiKey, int(organizationID)); err != nil {
		panic(err)
	}

	return &api
}

func getStripe(ctx context.Context) *stripeClient.API {
	httpClient := getHttpClient()
	stripeKey := os.Getenv("STRIPEKEY")

	sc := stripeClient.New(stripeKey, stripe.NewBackends(httpClient))

	return sc
}

func errorHandling(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Sprintln("error occured: ", err)
	fmt.Fprint(w, err)
	return
}
