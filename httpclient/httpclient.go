package httpclient

import (
	"log"
	"mime/multipart"
	"net/http"
)

type ImageHttpClient struct {
	url    string
	apiKey string
	client *http.Client
}

func NewImageHttpClient(url string, apiKey string) *ImageHttpClient {
	return &ImageHttpClient{
		url:    url,
		apiKey: apiKey,
		client: &http.Client{},
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

func (h ImageHttpClient) UploadImage(clientId string, image multipart.File, contentType string) (*http.Response, error) {
	log.Println("Uploading image to image service with for client id: " + clientId)
	req, err := http.NewRequest("POST", h.url+"/upload/"+clientId, image)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", h.apiKey)
	req.Header.Add("Content-Type", contentType)
	return h.client.Do(req)
}
