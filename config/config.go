package config

import (
	"os"
)

var (
	BusConnectionString = os.Getenv("AZURE_SERVICEBUS_CONNECTION_STRING")
	PublishQueueName    = os.Getenv("AZURE_PUBLISH_QUEUE_NAME")
	ReceiveQueueName    = os.Getenv("AZURE_RECEIVE_QUEUE_NAME")
	ImageServiceUrl     = os.Getenv("IMAGE_SERVICE_URL")
	ImageServiceApiKey  = os.Getenv("IMAGE_SERVICE_API_KEY")
)
