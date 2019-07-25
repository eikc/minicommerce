package http

import (
	"bytes"
	"encoding/json"
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

func TestGetAllProducts(t *testing.T) {
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockProductRepository(ctrl)

			server := Server{
				productRepository: repo,
				router:            httprouter.New(),
			}
			server.routes()

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

func TestGetProductByID(t *testing.T) {
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockProductRepository(ctrl)

			server := Server{
				productRepository: repo,
				router:            httprouter.New(),
			}
			server.routes()

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

func TestPostProduct(t *testing.T) {
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
		err     error
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
			err: nil,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockProductRepository(ctrl)
			mockDownloadables := mocks.NewMockDownloadableRepository(ctrl)

			server := Server{
				productRepository:      repo,
				downloadableRepository: mockDownloadables,
				router:                 httprouter.New(),
			}
			server.routes()

			for _, d := range tC.request.Product.Downloadables {
				item := minicommerce.Downloadable{
					ID:       d.ID,
					Name:     d.ID,
					Location: d.ID,
				}

				mockDownloadables.EXPECT().Get(gomock.Any(), d.ID).Times(1).Return(&item, nil)
			}

			repo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(1)

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
			cupaloy.SnapshotT(t, resp)
		})
	}
}

func TestPutProduct(t *testing.T) {
	testCases := []struct {
		desc    string
		request string
		err     error
	}{
		{
			desc: "Put product will work correctly",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {

		})
	}
}
