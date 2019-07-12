package storage

import (
	"context"

	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/gcp"
)

// Storage is the interface to the blob storage
type Storage struct {
	Client *gcp.HTTPClient
	Bucket string
}

// NewStorage creates the storage struct with all the needed dependencies
func NewStorage(client *gcp.HTTPClient, bucket string) *Storage {
	return &Storage{client, bucket}
}

// Read gets an object from the cloud storage
func (s *Storage) Read(ctx context.Context, location string) ([]byte, error) {
	b, err := gcsblob.OpenBucket(ctx, s.Client, s.Bucket, nil)
	if err != nil {
		return nil, err
	}
	defer b.Close()

	f, err := b.ReadAll(ctx, location)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// Write adds an new object to the cloud storage
func (s *Storage) Write(ctx context.Context, location string, file []byte) error {
	b, err := gcsblob.OpenBucket(ctx, s.Client, s.Bucket, nil)
	if err != nil {
		return err
	}
	defer b.Close()

	w, err := b.NewWriter(ctx, location, nil)
	if err != nil {
		return err
	}

	_, err = w.Write(file)
	if err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return nil
}

// Delete deletes an object from the cloud storage
func (s *Storage) Delete(ctx context.Context, location string) error {
	b, err := gcsblob.OpenBucket(ctx, s.Client, s.Bucket, nil)
	if err != nil {
		return err
	}

	if err := b.Delete(ctx, location); err != nil {
		return err
	}

	return nil
}
