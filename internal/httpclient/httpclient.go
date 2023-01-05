package httpclient

import (
	"log"
	"net/http"
	"time"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ImageHttpClient struct {
	url    string
	apiKey string
	client httpClient
}

func NewImageHttpClient(url string, apiKey string) *ImageHttpClient {
	return &ImageHttpClient{
		url:    url,
		apiKey: apiKey,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (h ImageHttpClient) GetImageFromQueue(uuid string) (*http.Response, error) {
	log.Println("Getting image from image service with uuid: " + uuid)
	req, err := http.NewRequest("GET", h.url+"/download/"+uuid, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", h.apiKey)
	return h.client.Do(req)
}
