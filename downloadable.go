package minicommerce

import (
	"context"
	"io"
)

// Downloadable is the location of a downloadable digital product uploaded somewhere to google cloud storage
type Downloadable struct {
	ID       string `firestore:"-"`
	Name     string `firestore:"name,omitempty"`
	Location string `firestore:"location,omitempty"`
}

// DownloadableReader ...
type DownloadableReader interface {
	Get(ctx context.Context, location string) (io.ReadCloser, error)
	GetAll(ctx context.Context) ([]Downloadable, error)
}

// DownloadableWriter ..
type DownloadableWriter interface {
	Create(ctx context.Context, location string, r io.Reader) error
}

// DownloadableDeleter ...
type DownloadableDeleter interface {
	Delete(ctx context.Context, location string) error
}

// DownloadableRepository ...
type DownloadableRepository interface {
	DownloadableReader
	DownloadableWriter
	DownloadableDeleter
}
