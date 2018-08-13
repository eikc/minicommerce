package main

import (
	"context"
	"os"
	"strconv"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func getSettings(ctx context.Context) Settings {
	key := datastore.NewKey(ctx, "settings", "settings", 0, nil)

	var s Settings
	datastore.Get(ctx, key, &s)

	return s
}

func initDevelopmentSettings(ctx context.Context) {
	if !appengine.IsDevAppServer() {
		return
	}

	orgId, _ := strconv.ParseInt(os.Getenv("CLIENTORGANIZATIONID"), 10, 64)

	settings := Settings{
		APIKey:                 os.Getenv("CLIENTAPIKEY"),
		ClientKey:              os.Getenv("CLIENTKEY"),
		ClientSecret:           os.Getenv("CLIENTSECRET"),
		InstagramToken:         "",
		OrganizationID:         int(orgId),
		StripeKey:              os.Getenv("STRIPEKEY"),
		StripeWebhookSignature: "",
	}

	key := datastore.NewKey(ctx, "settings", "settings", 0, nil)
	datastore.Put(ctx, key, &settings)
}
