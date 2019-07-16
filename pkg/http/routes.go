package http

import (
	"net/http"
)

func (s *Server) routes() {
	s.router.Handle(http.MethodGet, "/api/downloadables", s.getAllDownloadables())
	s.router.Handle(http.MethodPost, "/api/downloadbles", s.postDownloadables())

}
