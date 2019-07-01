package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

type MetaResponse struct {
	Meta *Meta
}

type UserResponse struct {
	User []User `json:"data"`
}

// User Object. Note that user objects are not always fully returned.
// Be sure to see the descriptions on the instagram documentation for any given endpoint.
type User struct {
	Images Images
	Link   string
}

type Images struct {
	Thumbnail          Image
	LowResolution      Image `json:"low_resolution"`
	StandardResolution Image `json:"standard_resolution"`
}

type Image struct {
	Width  int
	Height int
	Url    string
}

// Instagram User Counts object. Returned on User objects
type UserCounts struct {
	Media      int64
	Follows    int64
	FollowedBy int64 `json:"followed_by"`
}

type Pagination struct {
	NextUrl   string `json:"next_url"`
	NextMaxId string `json:"next_max_id"`

	// Used only on GetTagRecentMedia()
	NextMaxTagId string `json:"next_max_tag_id"`
	// Used only on GetTagRecentMedia()
	MinTagId string `json:"min_tag_id"`
}

type Meta struct {
	Code         int
	ErrorType    string `json:"error_type"`
	ErrorMessage string `json:"error_message"`
}

func getProfile(token string, httpClient *http.Client) (*UserResponse, error) {
	url := fmt.Sprint("https://api.instagram.com/v1/users/self/media/recent/?access_token=", token)

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userResp UserResponse
	if err := json.Unmarshal(b, &userResp); err != nil {
		return nil, err
	}

	return &userResp, nil
}

func instagram(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	client := getHttpClient()
	instragramToken := os.Getenv("instagram")

	url := fmt.Sprint("https://api.instagram.com/v1/users/self/media/recent/?access_token=", instragramToken)

	resp, err := client.Get(url)
	if err != nil {
		errorHandling(w, err)
		return
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorHandling(w, err)
		return
	}

	var userResp UserResponse
	if err := json.Unmarshal(b, &userResp); err != nil {
		errorHandling(w, err)
		return
	}

	userResp.User = userResp.User[:6]

	json, err := json.Marshal(userResp.User)
	if err != nil {
		errorHandling(w, err)
		return
	}

	w.Header().Add("Content-Type", "Application/json")
	fmt.Fprint(w, string(json))
}
