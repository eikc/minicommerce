package http

import (
	"net/http"

	"github.com/gofrs/uuid"

	"github.com/eikc/minicommerce"

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
		ctx := r.Context()
		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Should there be some kind of file validation ??
		// Can't figure it out - since there could be multiple usecases for downloadable content
		// Video, pdf's, executables, etc etc..

		ID, err := uuid.NewV4()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		downloadable := minicommerce.Downloadable{
			ID:       ID.String(),
			Name:     handler.Filename,
			Location: handler.Filename,
		}

		if err := s.downloadableRepository.Create(ctx, &downloadable); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := s.storage.Write(ctx, handler.Filename, file); err != nil {
			s.downloadableRepository.Delete(ctx, ID.String())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := response{
			ID:       downloadable.ID,
			Name:     downloadable.Name,
			Location: downloadable.Location,
		}

		sendJSON(w, 200, resp)
	}
}
