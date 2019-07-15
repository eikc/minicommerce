package firestore

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/eikc/minicommerce"
)

const downloadableCollection = "downloadables"

// DownloadableService handles data communication between firestore and the application
type DownloadableService struct {
	client *firestore.Client
}

// NewDownloadableService will construct the downlodable service correctly
func NewDownloadableService(client *firestore.Client) *DownloadableService {
	return &DownloadableService{client}
}

// GenerateID will generate a document ID that can be used to create documents based on
func (d *DownloadableService) GenerateID() string {
	return d.client.Collection(downloadableCollection).NewDoc().ID
}

// Get will return a downloadable document based on the id that is given
func (d *DownloadableService) Get(ctx context.Context, id string) (*minicommerce.Downloadable, error) {
	docRef := d.client.Collection(downloadableCollection).Doc(id)
	snapshot, err := docRef.Get(ctx)
	if err != nil {
		return nil, err
	}

	if !snapshot.Exists() {
		return nil, &DocumentNotFoundError{fmt.Sprintf("%s/%s", downloadableCollection, id)}
	}

	downloadable := minicommerce.Downloadable{
		ID: id,
	}

	if err = snapshot.DataTo(&downloadable); err != nil {
		return nil, err
	}

	var notDeleted int64
	if downloadable.Deleted != notDeleted {
		return nil, &DocumentNotFoundError{fmt.Sprintf("%s/%s", downloadableCollection, id)}
	}

	return &downloadable, nil
}

// GetAll will...
func (d *DownloadableService) GetAll() ([]minicommerce.Downloadable, error) {
	return nil, nil
}

// Create will..
func (d *DownloadableService) Create(ctx context.Context, downloadable *minicommerce.Downloadable) error {
	docRef := d.client.Collection(downloadableCollection).Doc(downloadable.ID)
	_, err := docRef.Create(ctx, downloadable)

	if err != nil {
		return err
	}

	return nil
}

// Delete will...
func (d *DownloadableService) Delete(ctx context.Context, id string) error {
	docRef := d.client.Collection(downloadableCollection).Doc(id)

	now := time.Now().UTC().Unix()
	_, err := docRef.Update(ctx, []firestore.Update{{Path: "deleted", Value: now}})
	if err != nil {
		return err
	}

	return nil
}
