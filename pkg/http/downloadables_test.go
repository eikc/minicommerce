package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"

	"github.com/eikc/minicommerce"
	"github.com/julienschmidt/httprouter"
)

type fakeDownloadableRepository struct {
	data []minicommerce.Downloadable
	err  error
}

func (f *fakeDownloadableRepository) Get(ctx context.Context, id string) (*minicommerce.Downloadable, error) {
	var downloadable minicommerce.Downloadable
	for _, d := range f.data {
		if d.ID == id {
			downloadable = d
			break
		}
	}

	return &downloadable, f.err
}

func (f *fakeDownloadableRepository) GetAll(ctx context.Context) ([]minicommerce.Downloadable, error) {
	return f.data, f.err
}

func (f *fakeDownloadableRepository) Create(ctx context.Context, downloadable *minicommerce.Downloadable) error {
	f.data = append(f.data, *downloadable)
	return f.err
}

func (f *fakeDownloadableRepository) Delete(ctx context.Context, id string) error {
	return f.err
}

func TestGetDownloadables(t *testing.T) {
	testCases := []struct {
		desc string
		err  error
		data []minicommerce.Downloadable
	}{
		{
			desc: "testing that the correct response is an collection of downloadables",
			err:  nil,
			data: []minicommerce.Downloadable{
				{
					ID:       "1",
					Name:     "some.pdf",
					Location: "some.pdf",
				},
				{
					ID:       "2",
					Name:     "some-2.pdf",
					Location: "some-2.pdf",
				},
			},
		},
		{
			desc: "testing error response",
			err:  fmt.Errorf("When an error occurs, internal server error is thrown"),
			data: nil,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			server := Server{
				downloadableRepository: &fakeDownloadableRepository{
					data: tC.data,
					err:  tC.err,
				},
				router: httprouter.New(),
			}
			server.routes()

			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "/api/downloadables", nil)
			if err != nil {
				t.Error(err.Error())
			}

			server.router.ServeHTTP(recorder, req)

			if status := recorder.Code; status != http.StatusOK {
				if tC.err != nil {
					resp := struct {
						Code int
						Err  string
					}{
						Code: recorder.Code,
						Err:  recorder.Body.String(),
					}

					cupaloy.SnapshotT(t, resp)
					return
				}

				t.Errorf("handler returned incorrect status")
			}

			var response struct {
				Collection []struct {
					ID   string
					Name string
				}
			}

			decoder := json.NewDecoder(recorder.Body)
			decoder.Decode(&response)

			cupaloy.SnapshotT(t, response)
		})
	}
}
