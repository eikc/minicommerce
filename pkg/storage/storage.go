package storage

import (
	"context"
	"io"

	"gocloud.dev/blob"

	// Enables the google cloud storage SDK
	_ "gocloud.dev/blob/gcsblob"
)

// Storage is the interface to the blob storage
type Storage struct {
	BucketURL string
}

// NewStorage creates the storage struct with all the needed dependencies
func NewStorage(bucketURL string) *Storage {
	return &Storage{bucketURL}
}

// Read gets an object from the cloud storage
func (s *Storage) Read(ctx context.Context, location string) (io.ReadCloser, error) {
	b, err := blob.OpenBucket(ctx, s.BucketURL)
	if err != nil {
		return nil, err
	}
	defer b.Close()

	r, err := b.NewReader(ctx, location, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Write adds an new object to the cloud storage
func (s *Storage) Write(ctx context.Context, location string, r io.Reader) error {
	b, err := blob.OpenBucket(ctx, s.BucketURL)
	if err != nil {
		return err
	}
	defer b.Close()

	w, err := b.NewWriter(ctx, location, nil)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, r)
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
	b, err := blob.OpenBucket(ctx, s.BucketURL)
	if err != nil {
		return err
	}

	if err := b.Delete(ctx, location); err != nil {
		return err
	}

	return nil
}
