package http

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (s *Server) getAllDownloadables() httprouter.Handle {
	type downloadableItem struct {
		ID   string `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
	}

	type response struct {
		Collection []downloadableItem `json:"collection,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()
		downloadables, err := s.downloadableRepository.GetAll(ctx)
		if err != nil {
			http.NotFound(w, r)
		}

		var resp response
		for _, downloadable := range downloadables {
			d := downloadableItem{
				ID:   downloadable.ID,
				Name: downloadable.Name,
			}

			resp.Collection = append(resp.Collection, d)
		}

		sendJSON(w, http.StatusOK, resp)
	}
}

func (s *Server) postDownloadables() httprouter.Handle {
	type response struct {
		ID       string `json:"id,omitempty"`
		Name     string `json:"name,omitempty"`
		Location string `json:"location,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	}
}
