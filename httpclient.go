package main

import (
	"net/http"
	"encoding/json"
)

type httpClient struct {
	client *http.Client
}


func NewHttpClient() *httpClient {
	return &httpClient{
		client: &http.Client{},
	}
}

func (client *httpClient) JSON(url string, value interface{}) error {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&value)
	if err != nil {
		return err
	}
	return err
}
