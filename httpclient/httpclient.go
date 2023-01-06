package httpclient

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
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

func (h ImageHttpClient) UploadImage(clientId string, imageFile *multipart.FileHeader, contentType string) (*http.Response, error) {
	log.Println("Uploading image to image service with for client id: " + clientId)

	image, err := imageFile.Open()
	if err != nil {
		return nil, err
	}
	defer image.Close()

	formData := url.Values{}
	formData.Add("file", imageFile.Filename)
	file, err := imageFile.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()
	buf := new(bytes.Buffer)
	io.Copy(buf, file)
	formData.Add("file", buf.String())

	req, err := http.NewRequest("POST", h.url+"/upload/"+clientId, strings.NewReader(formData.Encode()))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "multipart/form-data")
	req.Header.Add("Authorization", h.apiKey)
	return h.client.Do(req)
}
