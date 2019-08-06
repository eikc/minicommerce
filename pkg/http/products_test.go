package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eikc/minicommerce/pkg/firestore"

	"github.com/bradleyjkemp/cupaloy/v2"

	"github.com/eikc/minicommerce"

	"github.com/eikc/minicommerce/pkg/mocks"
	"github.com/julienschmidt/httprouter"

	"github.com/golang/mock/gomock"
)

func setupProductHTTPServer(t *testing.T) (*Server, *mocks.MockProductRepository,
	*mocks.MockDownloadableRepository,
	*mocks.MockTimeService,
	*mocks.MockIDGenerator,
	func()) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockProductRepository(ctrl)
	dRepo := mocks.NewMockDownloadableRepository(ctrl)
	time := mocks.NewMockTimeService(ctrl)
	uuidGenerator := mocks.NewMockIDGenerator(ctrl)

	server := Server{
		productRepository:      repo,
		downloadableRepository: dRepo,
		timeService:            time,
		idGenerator:            uuidGenerator,
		router:                 httprouter.New(),
	}
	server.routes()

	return &server, repo, dRepo, time, uuidGenerator, func() {
		ctrl.Finish()
	}
}

func TestProducts_GetAllProducts(t *testing.T) {
	testCases := []struct {
		desc     string
		products []minicommerce.Product
	}{
		{
			desc: "Get all will return the collection of products",
			products: []minicommerce.Product{
				{
					ID:          "Product-one",
					Created:     1,
					Updated:     2,
					Type:        minicommerce.ProductTypeDigital,
					Name:        "Test product one",
					Description: "This is a test product for a unit test",
					Price:       15000,
					Active:      true,
					Downloadable: []minicommerce.Downloadable{
						{
							ID:       "Testing-downloadable",
							Name:     "Coding cookbook for pro's",
							Location: "coding-cookbook.pdf",
						},
					},
				},
				{
					ID:           "Product-two",
					Created:      1,
					Updated:      2,
					Type:         minicommerce.ProductTypeLink,
					Name:         "Test product two",
					Description:  "Testing the product as linkable",
					Price:        15000,
					Active:       true,
					URL:          "https://some-url-to-the-linkable-product",
					Downloadable: []minicommerce.Downloadable{},
				},
			},
		},
		{
			desc:     "When no products exist, it will return an empty array as response",
			products: []minicommerce.Product{},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			server, repo, _, _, _, f := setupProductHTTPServer(t)
			defer f()

			repo.EXPECT().GetAll(gomock.Any()).Times(1).Return(tC.products, nil)

			recorder := httptest.NewRecorder()
			r, err := http.NewRequest(http.MethodGet, "/api/products", nil)
			if err != nil {
				t.Error(err.Error())
			}

			server.router.ServeHTTP(recorder, r)

			resp := struct {
				status int
				body   string
			}{
				status: recorder.Code,
				body:   recorder.Body.String(),
			}

			cupaloy.SnapshotT(t, resp)
		})
	}
}

func TestProducts_GetProductByID(t *testing.T) {
	testCases := []struct {
		desc    string
		id      string
		product *minicommerce.Product
		err     error
	}{
		{
			desc: "Getting a product by ID will return the correct product",
			id:   "product-one",
			product: &minicommerce.Product{
				ID:          "product-one",
				Created:     1,
				Updated:     2,
				Type:        minicommerce.ProductTypeDigital,
				Name:        "testing getting product",
				Description: "testing getting product by id",
				Price:       15000,
				Active:      true,
				Downloadable: []minicommerce.Downloadable{
					{
						ID:       "testing-with-downloadable",
						Name:     "some-pdf.pdf",
						Location: "somewhere/some.pdf",
					},
				},
			},
		},
		{
			desc:    "When no product exists, it will return 404",
			id:      "does-not-exist",
			product: nil,
			err:     &firestore.DocumentNotFoundError{},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			server, repo, _, _, _, finalize := setupProductHTTPServer(t)
			defer finalize()

			repo.EXPECT().Get(gomock.Any(), tC.id).Times(1).Return(tC.product, tC.err)

			recorder := httptest.NewRecorder()
			r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/products/%s", tC.id), nil)
			if err != nil {
				t.Error(err.Error())
			}

			server.router.ServeHTTP(recorder, r)

			resp := struct {
				status int
				body   string
			}{
				status: recorder.Code,
				body:   recorder.Body.String(),
			}

			cupaloy.SnapshotT(t, resp)
		})
	}
}

func TestProducts_PostProductResponses(t *testing.T) {
	type downloadables struct {
		ID string `json:"id,omitempty"`
	}
	type product struct {
		Type          minicommerce.ProductType `json:"type,omitempty"`
		Name          string                   `json:"name,omitempty"`
		Description   string                   `json:"description,omitempty"`
		Price         int64                    `json:"price,omitempty"`
		Active        bool                     `json:"active,omitempty"`
		URL           string                   `json:"url,omitempty"`
		Downloadables []downloadables          `json:"downloadables,omitempty"`
	}
	type request struct {
		Product product `json:"product"`
	}

	type response struct {
		Type          minicommerce.ProductType `json:"type"`
		Name          string                   `json:"name"`
		Description   string                   `json:"description"`
		Price         int64                    `json:"price"`
		Active        bool                     `json:"active"`
		URL           string                   `json:"url"`
		Downloadables []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Location string `json:"location"`
		} `json:"downloadables"`
	}

	testCases := []struct {
		desc    string
		request request
	}{
		{
			desc: "Post product will work correctly",
			request: request{
				Product: product{
					Type:        minicommerce.ProductTypeDigital,
					Name:        "testing a product create",
					Description: "testing a product create description",
					Price:       20000,
					Active:      true,
					Downloadables: []downloadables{
						{
							ID: "testing-testing",
						},
					},
				},
			},
		},
		{
			desc: "Post product will work with no downloadables",
			request: request{
				Product: product{
					Type:        minicommerce.ProductTypeLink,
					Name:        "testing a product create",
					Description: "testing a product create description",
					Price:       20000,
					Active:      true,
					URL:         "testing-url-thingie",
				},
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			server, repo, mockDownloadables, time, idgenerator, finalize := setupProductHTTPServer(t)
			defer finalize()

			for _, d := range tC.request.Product.Downloadables {
				item := minicommerce.Downloadable{
					ID:       d.ID,
					Name:     d.ID,
					Location: d.ID,
				}

				mockDownloadables.EXPECT().Get(gomock.Any(), d.ID).Times(1).Return(&item, nil)
			}

			repo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(1).Return(nil)
			time.EXPECT().Now().Times(1).Return(int64(123321))
			idgenerator.EXPECT().New().Times(1).Return("123-321-123-321", nil)

			recorder := httptest.NewRecorder()
			requestByte, _ := json.Marshal(tC.request)
			requestReader := bytes.NewReader(requestByte)
			r, err := http.NewRequest(http.MethodPost, "/api/products", requestReader)
			if err != nil {
				t.Error(err.Error())
			}
			server.router.ServeHTTP(recorder, r)

			var resp response
			json.Unmarshal(recorder.Body.Bytes(), &resp)

			result := struct {
				code int
				resp response
			}{
				code: recorder.Code,
				resp: resp,
			}

			cupaloy.SnapshotT(t, result)
		})
	}
}

func TestProducts_PostProductErrors(t *testing.T) {
	type downloadable struct {
		ID string `json:"id,omitempty"`
	}
	type product struct {
		Downloadables []downloadable `json:"downloadables,omitempty"`
	}
	type request struct {
		Product product `json:"product,omitempty"`
	}

	testCases := []struct {
		desc            string
		request         request
		err             error
		downloadableErr error
	}{
		{
			desc: "if an downloadable does not exist, it will return an http 404",
			request: request{
				Product: product{
					Downloadables: []downloadable{
						{
							ID: "something-that-does-not-exist",
						},
					},
				},
			},
			downloadableErr: errors.New("not found"),
		},
		{
			desc: "When the repository fails, we return an http 500",
			request: request{
				Product: product{
					Downloadables: []downloadable{
						{
							ID: "something-that-does-not-exist",
						},
					},
				},
			},
			err: errors.New("some test error occurred"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			server, repo, mockDownloadables, time, idgenerator, finalize := setupProductHTTPServer(t)
			defer finalize()

			time.EXPECT().Now().Times(1).Return(int64(123321))
			idgenerator.EXPECT().New().Times(1).Return("123-321-123-321", nil)

			for _, d := range tC.request.Product.Downloadables {
				item := minicommerce.Downloadable{
					ID:       d.ID,
					Name:     d.ID,
					Location: d.ID,
				}

				mockDownloadables.EXPECT().Get(gomock.Any(), d.ID).Times(1).Return(&item, tC.downloadableErr)
			}

			if tC.err != nil {
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(1).Return(tC.err)
			}

			recorder := httptest.NewRecorder()
			requestByte, _ := json.Marshal(tC.request)
			requestReader := bytes.NewReader(requestByte)
			r, err := http.NewRequest(http.MethodPost, "/api/products", requestReader)
			if err != nil {
				t.Error(err.Error())
			}
			server.router.ServeHTTP(recorder, r)

			result := struct {
				code int
				body string
			}{
				code: recorder.Code,
				body: recorder.Body.String(),
			}

			cupaloy.SnapshotT(t, result)
		})
	}
}

func TestProducts_PostProductRepositoryInputs(t *testing.T) {
	type downloadable struct {
		ID string `json:"id"`
	}

	type product struct {
		Type          string         `json:"type"`
		Name          string         `json:"name"`
		Description   string         `json:"description"`
		Price         int64          `json:"price"`
		Active        bool           `json:"active"`
		URL           string         `json:"url"`
		Downloadables []downloadable `json:"downloadables,omitempty"`
	}

	type request struct {
		Product product `json:"product"`
	}

	testCases := []struct {
		desc    string
		request request
	}{
		{
			desc: "The repository will insert a correct product with downloadables",
			request: request{
				Product: product{
					Type:        string(minicommerce.ProductTypeDigital),
					Name:        "testing digital product insertion",
					Description: "testing repository insertion",
					Price:       25000,
					Active:      true,
					Downloadables: []downloadable{
						{
							ID: "testing-digital-product-insertion",
						},
					},
				},
			},
		},
		{
			desc: "The repository input will be correct with no downloadables",
			request: request{
				Product: product{
					Type:          string(minicommerce.ProductTypeLink),
					Name:          "testing digital product insertion",
					Description:   "testing repository insertion",
					Price:         25000,
					Active:        true,
					Downloadables: nil,
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			server, repo, mockDownloadables, time, idgenerator, finalize := setupProductHTTPServer(t)
			defer finalize()

			time.EXPECT().Now().Times(1).Return(int64(123321))
			idgenerator.EXPECT().New().Times(1).Return("123-321-123-321", nil)

			for _, d := range tC.request.Product.Downloadables {
				item := minicommerce.Downloadable{
					ID:       d.ID,
					Name:     d.ID,
					Location: d.ID,
				}

				mockDownloadables.EXPECT().Get(gomock.Any(), d.ID).Times(1).Return(&item, nil)
			}

			var captured minicommerce.Product
			repo.EXPECT().Create(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, p *minicommerce.Product) {
				captured = *p
			}).Times(1).Return(nil)

			recorder := httptest.NewRecorder()
			requestByte, _ := json.Marshal(tC.request)
			requestReader := bytes.NewReader(requestByte)
			r, err := http.NewRequest(http.MethodPost, "/api/products", requestReader)
			if err != nil {
				t.Error(err.Error())
			}
			server.router.ServeHTTP(recorder, r)

			cupaloy.SnapshotT(t, captured)
		})
	}
}
