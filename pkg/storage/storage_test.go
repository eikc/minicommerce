package storage

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped")
	}

	ctx := context.Background()
	storage := NewStorage("gs://minicommerce_testing_123")

	if err := storage.Write(ctx, "testing.txt", strings.NewReader("hello world")); err != nil {
		t.Errorf(err.Error())
	}
	defer storage.Delete(ctx, "testing.txt")

	r, err := storage.Read(ctx, "testing.txt")
	if err != nil {
		t.Errorf(err.Error())
	}
	defer r.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Errorf(err.Error())
	}

	s := buf.String()
	if s != "hello world" {
		t.Errorf("the reader did not contain the correct text string")
	}
}
