package http

import (
	"net/http"
)

func (s *Server) routes() {
	// Downloadables
	s.router.Handle(http.MethodGet, "/api/downloadables", s.getAllDownloadables())
	s.router.Handle(http.MethodPost, "/api/downloadables", s.postDownloadables())

	// Products
	s.router.Handle(http.MethodGet, "/api/products", s.getAllProducts())
	s.router.Handle(http.MethodGet, "/api/products/:id", s.getProductByID())
	s.router.Handle(http.MethodPost, "/api/products", s.postProduct())
}
