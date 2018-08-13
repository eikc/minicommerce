package main

import (
	"context"
	"fmt"
	"net/http"

	dinero "github.com/eikc/dinero-go"
	stripe "github.com/stripe/stripe-go"
	stripeClient "github.com/stripe/stripe-go/client"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

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

func errorHandling(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Sprintln("error occured: ", err)
	fmt.Fprint(w, err)
	return
}
