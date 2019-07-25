package http

import (
	"net/http"

	"github.com/eikc/minicommerce/pkg/firestore"

	"github.com/eikc/minicommerce"

	"github.com/julienschmidt/httprouter"
)

func (s *Server) getAllProducts() httprouter.Handle {

	type response struct {
		Collection []minicommerce.Product `json:"collection"`
	}

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()
		products, err := s.productRepository.GetAll(ctx)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		resp := response{
			Collection: products,
		}

		sendJSON(w, 200, resp)
	}
}

func (s *Server) getProductByID() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		ctx := r.Context()
		id := params.ByName("id")
		product, err := s.productRepository.Get(ctx, id)
		if err != nil {
			switch err.(type) {
			case *firestore.DocumentNotFoundError:
				http.Error(w, "product not found", 404)
				return
			default:
				http.Error(w, err.Error(), 500)
				return
			}

		}

		sendJSON(w, 200, product)
	}
}

func (s *Server) postProduct() httprouter.Handle {

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	}
}

func (s *Server) putProduct() httprouter.Handle {

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	}
}
