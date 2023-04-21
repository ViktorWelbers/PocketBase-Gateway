package handlers

import (
	"api-gateway/internal/messaging"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/models"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type imageQueueClient interface {
	ReceiveMessage(messageId string) string
	SendMessage(message messaging.Message) error
}

type httpClient interface {
	GetImageFromQueue(uuid string) (*http.Response, error)
}

func NewImageDownloadHandler(queueClient imageQueueClient, httpClient httpClient) *ImageDownload {
	return &ImageDownload{
		ImageQueueClient: queueClient,
		ImageHttpClient:  httpClient,
	}
}

type ImageDownload struct {
	ImageQueueClient imageQueueClient
	ImageHttpClient  httpClient
}

func (h *ImageDownload) Handle(c echo.Context) error {
	log.Print("Received request to retrieve image")
	messageId := c.QueryParam("message_id")

	newUuid := h.ImageQueueClient.ReceiveMessage(messageId)
	if newUuid == "" {
		return c.String(http.StatusAccepted, "Message not ready yet")
	}

	// Create the response object
	resp, err := h.ImageHttpClient.GetImageFromQueue(newUuid)
	defer resp.Body.Close()
	if err != nil {
		return apis.NewBadRequestError("Error when getting image from queue", nil)
	}

	log.Println(resp.Header)
	image, err := io.ReadAll(resp.Body)
	if err != nil {
		return apis.NewBadRequestError("Error when trying to get image from body", nil)
	}

	// Encode the response as Image
	return c.Blob(http.StatusOK, resp.Header.Get("Content-Type"), image)
}

func NewAuthenticationHandler() *Authentication {
	return &Authentication{}
}

type Authentication struct{}

func (h *Authentication) Handle(c echo.Context) error {
	authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
	if authRecord == nil {
		return apis.NewForbiddenError("Only auth records can access this endpoint", nil)
	}
	userName := authRecord.Username()

	log.Print("Received request to confirm login")
	return c.JSON(http.StatusOK, map[string]string{
		"client_id": userName,
	})
}

type MessageProducer struct {
	ImageQueueClient imageQueueClient
}

func NewMessageProducerHandler(queueClient imageQueueClient) *MessageProducer {
	return &MessageProducer{
		ImageQueueClient: queueClient,
	}
}

func (h *MessageProducer) Handle(c echo.Context) error {
	prompt := c.QueryParam("prompt")
	promptGuidance := c.QueryParam("prompt_guidance")
	strength := c.QueryParam("strength")
	uuid := c.QueryParam("uuid")
	modelType := c.QueryParam("model_type")
	if prompt == "" || promptGuidance == "" || strength == "" || uuid == "" {
		return apis.NewBadRequestError("Missing required query parameters", nil)
	}
	strengthFloat, _ := strconv.ParseFloat(strength, 32)
	promptGuidanceFloat, _ := strconv.ParseFloat(promptGuidance, 32)
	messageId := strconv.FormatInt(time.Now().Unix(), 10) + "-" + uuid
	// SendMessage the message to the queue using the query parameters
	message := messaging.Message{
		Prompt:         prompt,
		PromptGuidance: promptGuidanceFloat,
		Strength:       strengthFloat,
		Uuid:           uuid,
		MessageId:      messageId,
		ModelType:      modelType,
		Token:          c.Request().Header.Get("Authorization"),
	}
	err := h.ImageQueueClient.SendMessage(message)
	if err != nil {
		return apis.NewBadRequestError("Queue was not able to send message", err.Error())
	}
	// SendMessage Response
	return c.JSON(http.StatusOK, map[string]string{
		"messageId": messageId,
	})
}
