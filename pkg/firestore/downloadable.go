package firestore

import (
	"context"
	"fmt"

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

	return &downloadable, nil
}

// GetAll will get all non deleted downloadables from firestore
func (d *DownloadableService) GetAll(ctx context.Context) ([]minicommerce.Downloadable, error) {
	colRef := d.client.Collection(downloadableCollection)
	iter := colRef.Documents(ctx)
	docs, err := iter.GetAll()
	if err != nil {
		return nil, err
	}

	var collection []minicommerce.Downloadable
	for _, doc := range docs {
		var data minicommerce.Downloadable
		doc.DataTo(&data)
		data.ID = doc.Ref.ID
		collection = append(collection, data)
	}

	return collection, nil
}

// Create will create a documents in firestore with the given data, if the document ID exist it will fail
func (d *DownloadableService) Create(ctx context.Context, downloadable *minicommerce.Downloadable) error {
	docRef := d.client.Collection(downloadableCollection).Doc(downloadable.ID)
	_, err := docRef.Create(ctx, downloadable)

	if err != nil {
		return err
	}

	return nil
}

// Delete will remove a document from the firestore collection
func (d *DownloadableService) Delete(ctx context.Context, id string) error {
	docRef := d.client.Collection(downloadableCollection).Doc(id)
	_, err := docRef.Collection(downloadableCollection).Doc(id).Delete(ctx)
	if err != nil {
		return err
	}

	return nil
}
