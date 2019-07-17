package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/eikc/minicommerce/pkg/mocks"

	"github.com/golang/mock/gomock"

	"github.com/bradleyjkemp/cupaloy/v2"

	"github.com/eikc/minicommerce"
	"github.com/julienschmidt/httprouter"
)

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
			desc: "When no downloadables are available, it will be an empty collection response",
			err:  nil,
			data: []minicommerce.Downloadable{},
		},
		{
			desc: "testing error response",
			err:  fmt.Errorf("When an error occurs, internal server error is thrown"),
			data: nil,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockDownloadableRepository(ctrl)
			mockRepo.EXPECT().GetAll(gomock.Any()).Return(tC.data, tC.err).Times(1)

			server := Server{
				downloadableRepository: mockRepo,
				router:                 httprouter.New(),
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

func TestCreateDownloadable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockDownloadableRepository(ctrl)
	storage := mocks.NewMockStorage(ctrl)

	var capturedDownloadable struct {
		Name     string
		Location string
	}
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, d *minicommerce.Downloadable) {
		capturedDownloadable = struct {
			Name     string
			Location string
		}{
			Name:     d.Name,
			Location: d.Location,
		}
	}).Times(1)
	storage.EXPECT().Write(gomock.Any(), "simple.pdf", gomock.Any()).Times(1)

	server := Server{
		downloadableRepository: repo,
		storage:                storage,
		router:                 httprouter.New(),
	}
	server.routes()

	recorder := httptest.NewRecorder()
	r, err := newfileUploadRequest("/api/downloadables", "file", "./testfiles/simple.pdf")
	if err != nil {
		t.Error(err.Error())
	}

	server.router.ServeHTTP(recorder, r)

	cupaloy.SnapshotT(t, capturedDownloadable)
}

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}
