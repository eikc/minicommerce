package main

import (
	"context"
	"log"
)

func main() {
	ctx := context.Background()
	srv, err := NewServer(ctx, "gs://minicommerce_testing_123", "minicommerce-testing")
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Printf("Listening on port %s", "8080")
	log.Fatal(srv.Run("8080"))
}
