package firestore

import (
	"context"
	"testing"

	"github.com/bradleyjkemp/cupaloy"

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

	ID := "testing-get-1"

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

	cupaloy.SnapshotT(t, downloadable)
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

func TestDownloadableServiceDelete(t *testing.T) {
	ctx := context.Background()

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		t.Errorf(err.Error())
	}

	doc := client.Collection(downloadableCollection).NewDoc()
	defer func() {
		client.Collection(downloadableCollection).Doc(doc.ID).Delete(ctx)
		client.Close()
	}()

	data := minicommerce.Downloadable{
		ID:       doc.ID,
		Name:     "testing delete",
		Location: "somepdf.pdf",
	}

	_, err = doc.Set(ctx, data)
	if err != nil {
		t.Errorf(err.Error())
	}

	downloadableService := NewDownloadableService(client)

	if err = downloadableService.Delete(ctx, doc.ID); err != nil {
		t.Errorf(err.Error())
	}
}

func TestGetAllDownloadable(t *testing.T) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		t.Errorf(err.Error())
	}
	dd := []minicommerce.Downloadable{
		{ID: "test-1", Name: "test 1", Location: "one.pdf"},
		{ID: "test-2", Name: "test 2", Location: "two.pdf"},
		{ID: "test-3", Name: "test 3", Location: "three.pdf"},
	}

	defer func() {
		for _, d := range dd {
			client.Collection(downloadableCollection).Doc(d.ID).Delete(ctx)
		}
		client.Close()
	}()

	for _, d := range dd {
		docRef := client.Collection(downloadableCollection).Doc(d.ID)
		docRef.Set(ctx, d)
	}

	service := NewDownloadableService(client)

	downloadables, err := service.GetAll(ctx)
	if err != nil {
		t.Errorf(err.Error())
	}

	cupaloy.SnapshotT(t, downloadables)
}
