package main

import (
	"context"
	"log"
	"os"

	"github.com/eikc/minicommerce/pkg/storage"
)

func main() {
	ctx := context.Background()
	bucketURL := os.Getenv("bucketURL")
	projectID := os.Getenv("projectID")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv, err := NewServer(ctx, storage.BucketURL(bucketURL), projectID)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(srv.Run(port))
}
