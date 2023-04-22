package messaging

import (
	"context"
	"encoding/json"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"log"
)

type Receiver interface {
	Close(ctx context.Context) error
	ReceiveMessages(ctx context.Context, maxMessageCount int32, opts *azservicebus.ReceiveMessagesOptions) ([]*azservicebus.Message, error)
	CompleteMessage(ctx context.Context, message *azservicebus.Message, opts *azservicebus.CompleteMessageOptions) error
}

type QueueClient struct {
	AzureClient        *azservicebus.Client
	ConnectionString   string
	PublishQueueName   string
	ReceivingQueueName string
}

type Message struct {
	MessageId      string  `json:"message_id"`
	Uuid           string  `json:"uuid"`
	Prompt         string  `json:"prompt"`
	PromptGuidance float64 `json:"prompt_guidance"`
	Strength       float64 `json:"strength"`
	ModelType      string  `json:"model_type"`
	Token          string  `json:"token"`
}

func NewQueueClient(receivingQueueName, publishQueueName, connectingString string) *QueueClient {
	client, err := azservicebus.NewClientFromConnectionString(connectingString, nil)
	if err != nil {
		panic(err)
	}
	return &QueueClient{
		AzureClient:        client,
		ConnectionString:   connectingString,
		PublishQueueName:   publishQueueName,
		ReceivingQueueName: receivingQueueName,
	}
}

func (client *QueueClient) SendMessage(m Message) error {
	log.Print("Sending message to queue")
	sender, err := client.AzureClient.NewSender(client.PublishQueueName, nil)
	if err != nil {
		panic(err)
	}
	defer sender.Close(context.TODO())

	body, _ := json.Marshal(m)
	sbMessage := &azservicebus.Message{
		Body: body,
	}
	err = sender.SendMessage(context.TODO(), sbMessage, nil)
	return err
}

func (client *QueueClient) ReceiveMessage(messageId string) string {
	log.Print("Receiving message from queue")
	var result map[string]interface{}
	receiver, err := client.AzureClient.NewReceiverForQueue(client.ReceivingQueueName, nil)
	if err != nil {
		panic(err)
	}
	defer receiver.Close(context.TODO())

	messages, err := receiver.ReceiveMessages(context.TODO(), 100, nil)
	if err != nil {
		panic(err)
	}

	for _, message := range messages {

		body := message.Body
		err := json.Unmarshal(body, &result)

		if result["message_id"] == messageId {
			log.Print("Found message with id: ", messageId)
			err = receiver.CompleteMessage(context.TODO(), message, nil)
			if err != nil {
				panic(err)
			}
			return result["uuid_processed"].(string)
		}
		if err != nil {
			panic(err)
		}

	}
	return ""
}
