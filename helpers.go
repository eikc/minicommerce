package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	dinero "github.com/eikc/dinero-go"
	stripe "github.com/stripe/stripe-go"
	stripeClient "github.com/stripe/stripe-go/client"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

func getClient(ctx context.Context) *dineroAPI {
	apiKey := os.Getenv("CLIENTAPIKEY")
	clientKey := os.Getenv("CLIENTKEY")
	clientSecret :=os.Getenv("CLIENTSECRET")
	organizationID, _ := strconv.ParseInt(os.Getenv("CLIENTORGANIZATIONID"), 10, 64)


	httpClient := getHttpClient()
	dineroClient := dinero.NewClient(clientKey, clientSecret, httpClient)

	api := dineroAPI{
		API: dineroClient,
	}

	if err := api.Authorize(apiKey, int(organizationID)); err != nil {
		log.Criticalf(ctx, "Can't authorize with dinero, settings: %v - err: %v", s, err)
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
