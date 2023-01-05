package handlers

import (
	"api-gateway/httpclient"
	"api-gateway/messaging"
	"encoding/json"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Handler interface {
	Handle(c echo.Context) error
}

type MessageProducer struct {
	// The client to use for the handler
	QueueClient *messaging.QueueClient
}

type ImageDownload struct {
	// The client to use for the handler
	QueueClient     *messaging.QueueClient
	ImageHttpClient *httpclient.ImageHttpClient
}

type ImageUpload struct {
	// The client to use for the handler
	ImageHttpClient *httpclient.ImageHttpClient
}

func NewMessageProducerHandler(queueClient *messaging.QueueClient) Handler {
	return &MessageProducer{
		QueueClient: queueClient,
	}
}

func NewImageDownloadHandler(queueClient *messaging.QueueClient, httpClient *httpclient.ImageHttpClient) Handler {
	return &ImageDownload{
		QueueClient:     queueClient,
		ImageHttpClient: httpClient,
	}
}

func NewImageUploadHandler(httpClient *httpclient.ImageHttpClient) Handler {
	return &ImageUpload{
		ImageHttpClient: httpClient,
	}
}

func (h *ImageDownload) Handle(c echo.Context) error {
	//authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
	//if authRecord == nil {
	//	return apis.NewForbiddenError("Only auth records can access this endpoint", nil)
	//}
	log.Print("Received request to retrieve image")
	messageId := c.QueryParam("message_id")

	newUuid := messaging.ReceiveMessage(h.QueueClient, messageId)
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

func (h *ImageUpload) Handle(c echo.Context) error {
	var jsonData map[string]interface{}
	log.Println("Received request to upload image")
	imageFile, err := c.FormFile("file")
	if err != nil {
		return apis.NewBadRequestError("Error when trying to get image from body", err.Error())
	}
	image, err := imageFile.Open()
	if err != nil {
		return apis.NewBadRequestError("Error opening the file", err.Error())
	}
	defer image.Close()

	file := &os.File{}
	_, err = io.Copy(file, image)
	if err != nil {
		return apis.NewBadRequestError("Error reading the file", err.Error())
	}
	defer file.Close()

	//clientId := c.Get(apis.ContextAuthRecordKey).(*models.Record).Username()
	clientId := "test"
	resp, err := h.ImageHttpClient.UploadImage(clientId, file)
	if err != nil {
		return apis.NewBadRequestError("Error when uploading image via http client", err.Error())
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return apis.NewBadRequestError("Error when trying to get data from body", err.Error())
	}
	log.Println(resp.StatusCode)

	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		log.Println(err)
		return apis.NewBadRequestError("Error when trying to unmarshal data", err.Error())
	}
	return c.JSON(http.StatusOK, jsonData)
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
	// Send the message to the queue using the query parameters
	message := messaging.Message{
		Prompt:         prompt,
		PromptGuidance: promptGuidanceFloat,
		Strength:       strengthFloat,
		Uuid:           uuid,
		MessageId:      messageId,
		ModelType:      modelType,
	}
	err := message.Send(h.QueueClient)
	if err != nil {
		return apis.NewBadRequestError("Queue was not able to send message", err.Error())
	}
	// Send Response
	return c.JSON(http.StatusOK, map[string]string{
		"messageId": messageId,
	})
}
