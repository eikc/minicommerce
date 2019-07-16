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
		Collection []downloadableItem `json:"collection"`
	}

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()
		downloadables, err := s.downloadableRepository.GetAll(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := response{
			Collection: make([]downloadableItem, 0),
		}
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
