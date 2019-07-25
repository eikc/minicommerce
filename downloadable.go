package minicommerce

import (
	"context"
)

// Downloadable is the location of a downloadable digital product uploaded somewhere to google cloud storage
type Downloadable struct {
	ID       string `firestore:"-" json:"id"`
	Name     string `firestore:"name" json:"name"`
	Location string `firestore:"location" json:"location"`
}

// DownloadableReader ...
type DownloadableReader interface {
	Get(ctx context.Context, id string) (*Downloadable, error)
	GetAll(ctx context.Context) ([]Downloadable, error)
}

// DownloadableWriter ..
type DownloadableWriter interface {
	Create(ctx context.Context, downloadable *Downloadable) error
}

// DownloadableDeleter ...
type DownloadableDeleter interface {
	Delete(ctx context.Context, id string) error
}

// DownloadableRepository ...
type DownloadableRepository interface {
	DownloadableReader
	DownloadableWriter
	DownloadableDeleter
}
