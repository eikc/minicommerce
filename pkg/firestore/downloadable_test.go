package firestore

import (
	"context"
	"reflect"
	"testing"

	"github.com/eikc/minicommerce"

	"cloud.google.com/go/firestore"
)

const projectID = "minicommerce-testing"

func TestDownloadableServiceGet(t *testing.T) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		t.Errorf(err.Error())
	}

	ID := client.Collection(downloadableCollection).NewDoc().ID

	defer func() {
		client.Collection(downloadableCollection).Doc(ID).Delete(ctx)
		client.Close()
	}()

	document := &minicommerce.Downloadable{
		ID:       ID,
		Name:     "testing downloadable get",
		Location: "testing.pdf",
	}
	client.Collection(downloadableCollection).Doc(ID).Set(ctx, document)

	downloadableService := NewDownloadableService(client)

	downloadable, err := downloadableService.Get(ctx, ID)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !reflect.DeepEqual(*document, *downloadable) {
		t.Errorf("Downloadable documents does not match")
	}
}

func TestDownloadableServiceCreate(t *testing.T) {
	ctx := context.Background()

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		t.Errorf(err.Error())
	}
	ID := client.Collection(downloadableCollection).NewDoc().ID

	defer func() {
		client.Collection(downloadableCollection).Doc(ID).Delete(ctx)
		client.Close()
	}()

	downloadableService := NewDownloadableService(client)

	doc := minicommerce.Downloadable{
		ID:       ID,
		Name:     "test",
		Location: "some.pdf",
	}

	err = downloadableService.Create(ctx, &doc)
	if err != nil {
		t.Errorf(err.Error())
	}
}
