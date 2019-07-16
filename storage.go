package minicommerce

import (
	"context"
	"io"
)

// StorageWriter ..
type StorageWriter interface {
	Write(ctx context.Context, location string, r io.Reader) error
}

// StorageReader ...
type StorageReader interface {
	Read(ctx context.Context, location string) (io.ReadCloser, error)
}

// StorageDeleter ..
type StorageDeleter interface {
	Delete(ctx context.Context, location string) error
}

// Storage is the engine that can read,write,delete storage objects
type Storage interface {
	StorageWriter
	StorageReader
	StorageDeleter
}
