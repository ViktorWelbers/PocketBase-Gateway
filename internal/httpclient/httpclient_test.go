package httpclient

import (
	"errors"
	"net/http"
	"testing"
)

type mockHttpClient struct{}

func (c *mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	if req.URL.String() != "/download/mock_uuid" {
		return nil, errors.New("expected URL to be '/download/mock_uuid'")
	}
	if req.Header.Get("Authorization") != "mock_api_key" {
		return nil, errors.New("expected 'Authorization' header to be set to 'mock_api_key'")
	}
	return &http.Response{
		StatusCode: http.StatusOK,
	}, nil
}

func TestGetImageFromQueue(t *testing.T) {
	client := &ImageHttpClient{
		url:    "mock_url",
		apiKey: "mock_api_key",
		client: &mockHttpClient{},
	}

	response, err := client.GetImageFromQueue("mock_uuid")
	if err != nil {
		t.Errorf("Expected no error, but got '%v'", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, response.StatusCode)
	}
}
