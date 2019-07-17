//+build wireinject

package main

import (
	"context"

	"google.golang.org/api/option"

	"github.com/eikc/minicommerce"

	f "cloud.google.com/go/firestore"

	"github.com/eikc/minicommerce/pkg/firestore"
	"github.com/eikc/minicommerce/pkg/storage"
	"github.com/google/wire"

	"github.com/eikc/minicommerce/pkg/http"
)

// NewServer is using wire to construct the correct server struct
func NewServer(ctx context.Context, bucketURL storage.BucketURL, projectID string, opts ...option.ClientOption) (*http.Server, error) {

	wire.Build(
		http.NewServer,
		f.NewClient,
		storage.NewStorage,
		firestore.NewDownloadableService,
		wire.Bind(new(minicommerce.Storage), new(storage.Storage)),
		wire.Bind(new(minicommerce.DownloadableRepository), new(firestore.DownloadableService)))

	return &http.Server{}, nil
}
