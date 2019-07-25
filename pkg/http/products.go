package http

import (
	"net/http"
	"time"

	"github.com/gofrs/uuid"

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
	type request struct {
		Product struct {
			Type          minicommerce.ProductType `json:"type"`
			Name          string                   `json:"name"`
			Description   string                   `json:"description"`
			Price         int64                    `json:"price"`
			Active        bool                     `json:"active"`
			URL           string                   `json:"url"`
			Downloadables []struct {
				ID string `json:"id"`
			} `json:"downloadables"`
		} `json:"product"`
	}

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()
		id, err := uuid.NewV4()

		if r.Body == nil {
			http.Error(w, "Incorrect request", http.StatusBadRequest)
		}

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var request request
		if err := receiveJSON(r.Body, &request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		created := time.Now().Unix()

		product := minicommerce.Product{
			ID:          id.String(),
			Created:     created,
			Updated:     created,
			Type:        request.Product.Type,
			Name:        request.Product.Name,
			Description: request.Product.Description,
			Price:       request.Product.Price,
			Active:      request.Product.Active,
			URL:         request.Product.URL,
		}

		var downloadables []minicommerce.Downloadable
		for _, d := range request.Product.Downloadables {
			// this can be optimized by using firestore getAll Document refs
			downloadable, err := s.downloadableRepository.Get(ctx, d.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			downloadables = append(downloadables, *downloadable)
		}

		product.Downloadable = downloadables

		// some product validation would be nice to make sure we don't save anything stupid..
		// it's on the todo list
		if err := s.productRepository.Create(ctx, &product); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sendJSON(w, 200, product)
	}
}
