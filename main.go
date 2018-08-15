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
	router.GET("/downloads/:orderid", download)

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

func download(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	orderID := params.ByName("orderid")
	ctx := appengine.NewContext(r)
	stripeAPI := getStripe(ctx)

	order, err := stripeAPI.Orders.Get(orderID, nil)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, "404 page not found")
		return
	}

	w.WriteHeader(200)
	fmt.Fprintln(w, "ID is: ", order.ID)
}
