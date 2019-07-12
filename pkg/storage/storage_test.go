package storage

import (
	"context"
	"testing"
)

func TestStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped")
	}

	ctx := context.Background()
	storage := NewStorage("gs://minicommerce_testing")

	if err := storage.Write(ctx, "testing.txt", []byte("hello world")); err != nil {
		t.Errorf(err.Error())
	}

	b, err := storage.Read(ctx, "testing.txt")
	if err != nil {
		t.Errorf(err.Error())
	}

	if string(b) != "hello world" {
		t.Error("Could not read the correct document")
	}

	if err = storage.Delete(ctx, "testing.txt"); err != nil {
		t.Errorf(err.Error())
	}
}
