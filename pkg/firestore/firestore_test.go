package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
)

const projectID = "minicommerce-testing"

func cleanup(c *firestore.Client, collection, ID string) {
	c.Collection(collection).Doc(ID).Delete(context.Background())
	c.Close()
}
