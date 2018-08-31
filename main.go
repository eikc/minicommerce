package main

import (
	"fmt"
	"net/http"

	"github.com/rs/cors"
	"google.golang.org/appengine"

	"github.com/julienschmidt/httprouter"
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

	handler := cors.Default().Handler(router)

	http.Handle("/", handler)
	appengine.Main()
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "test")
}

func bootstrap(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := appengine.NewContext(r)
	initDevelopmentSettings(ctx)

	fmt.Fprint(w, "test")
}
