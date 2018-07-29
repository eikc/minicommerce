package main

import (
	"context"

	"google.golang.org/appengine/datastore"
)

func getSettings(ctx context.Context) Settings {
	key := datastore.NewKey(ctx, "settings", "settings", 0, nil)

	var s Settings
	datastore.Get(ctx, key, &s)

	return s
}
