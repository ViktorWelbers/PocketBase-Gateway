package handlers

import (
	"api-gateway/internal/messaging"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type SpyImageQueueClient struct {
	ReceiveCalls int
	SendCalls    int
}

func (s *SpyImageQueueClient) ReceiveMessage(_ string) string {
	s.ReceiveCalls++
	return "123"
}

func (s *SpyImageQueueClient) SendMessage(_ messaging.Message) error {
	s.SendCalls++
	return nil
}

type SpyHttpClient struct {
	Calls int
}

func (s *SpyHttpClient) GetImageFromQueue(_ string) (*http.Response, error) {
	s.Calls++
	return nil, nil
}

func TestAuthenticationHandlerHandleError(t *testing.T) {
	// given
	authentication := NewAuthenticationHandler()
	if authentication == nil {
		t.Error("Authentication handler is nil")
	}
	c := echo.New().NewContext(nil, nil)
	want := apis.NewForbiddenError("Only auth records can access this endpoint", nil)
	// when
	got := authentication.Handle(c)

	// then
	if got.Error() != want.Error() {
		t.Errorf("Authentication.Handle() = %v, want %v", got, want)
	}

}

func TestNewImageDownloadHandler(t *testing.T) {
	// given
	spyImageQueueClient := &SpyImageQueueClient{}
	spyHttpClient := &SpyHttpClient{}
	handler := NewImageDownloadHandler(spyImageQueueClient, spyHttpClient)
	context := echo.New().NewContext(nil, nil)
	context.QueryParams().Add("message_id", "123")

	// when
	err := handler.Handle(context)
	if err != nil {
		t.Error(err)
	}

	// then
	assert.Equal(t, spyHttpClient.Calls, 1)
	assert.Equal(t, spyImageQueueClient.ReceiveCalls, 1)
}

func TestMessageProducer_Handle(t *testing.T) {
	// given
	spyImageQueueClient := &SpyImageQueueClient{}
	handler := NewMessageProducerHandler(spyImageQueueClient)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/path?prompt=prompt&prompt_guidance=1.0&strength=1.0&uuid=uuid&model_type=model", strings.NewReader(""))
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// when
	err := handler.Handle(c)

	// then
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "1", c.QueryParam("strength"))
	assert.Equal(t, "1", c.QueryParam("prompt_guidance"))
	assert.Equal(t, "prompt", c.QueryParam("prompt"))
	assert.Equal(t, "uuid", c.QueryParam("uuid"))
	assert.Equal(t, "model", c.QueryParam("model_type"))
	assert.Equal(t, "Bearer token", c.Request().Header.Get("Authorization"))
	assert.Equal(t, 1, spyImageQueueClient.SendCalls)
	assert.Contains(t, rec.Body.String(), `"messageId":"`)
}
