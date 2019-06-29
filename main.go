package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
)

func main() {
	router := httprouter.New()
	router.GET("/", index)
	router.POST("/create", create())
	router.POST("/webhook", webhookReceiver())
	router.GET("/instagram", instagram)
	router.GET("/bootcamp", bootcamp)
	router.POST("/bootcamp", buyBootcamp)
	router.GET("/downloads/:orderid", download)
	router.GET("/downloads/:orderid/:sku", downloadV2)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	handler := cors.Default().Handler(router)

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), handler))
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "test")
}

func getHttpClient() *http.Client {
	return &http.Client{Timeout: time.Second * 60}
}

func isDevelopmentServer() bool {
	env := os.Getenv("ENVIRONMENT")

	if env == "production" {
		return false
	}

	return true
}
