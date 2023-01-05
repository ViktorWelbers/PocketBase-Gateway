package messaging

import (
	"context"
	"encoding/json"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"log"
)

func NewQueueSender(publishQueueName, connectingString string) *azservicebus.Sender {
	client, err := azservicebus.NewClientFromConnectionString(connectingString, nil)
	if err != nil {
		panic(err)
	}
	sender, err := client.NewSender(publishQueueName, nil)
	if err != nil {
		panic(err)
	}
	return sender
}

func NewQueueReceiver(subscribeQueueName, connectingString string) *azservicebus.Receiver {
	client, err := azservicebus.NewClientFromConnectionString(connectingString, nil)
	if err != nil {
		panic(err)
	}
	receiver, err := client.NewReceiverForQueue(subscribeQueueName, nil)
	if err != nil {
		panic(err)
	}
	return receiver
}

func NewQueueClient(publishQueueName, subscribeQueueName, connectingString string) *QueueClient {
	sender := NewQueueSender(publishQueueName, connectingString)
	receiver := NewQueueReceiver(subscribeQueueName, connectingString)
	return &QueueClient{
		receiver: receiver,
		sender:   sender,
	}
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

type QueueReceiver interface {
	Close(ctx context.Context) error
	ReceiveMessages(ctx context.Context, maxMessages int, options *azservicebus.ReceiveMessagesOptions) ([]*azservicebus.ReceivedMessage, error)
	CompleteMessage(ctx context.Context, message *azservicebus.ReceivedMessage, options *azservicebus.CompleteMessageOptions) error
}

type QueueSender interface {
	Close(ctx context.Context) error
	SendMessage(ctx context.Context, message *azservicebus.Message, opts *azservicebus.SendMessageOptions) error
}

type QueueClient struct {
	receiver QueueReceiver
	sender   QueueSender
}

func (c *QueueClient) Close() {
	_ = c.receiver.Close(context.Background())
	_ = c.sender.Close(context.Background())
}

func (c *QueueClient) SendMessage(m Message) error {
	log.Print("Sending message to queue")

	body, _ := json.Marshal(m)
	sbMessage := &azservicebus.Message{
		Body: body,
	}
	err := c.sender.SendMessage(context.TODO(), sbMessage, nil)
	return err
}

func (c *QueueClient) ReceiveMessage(messageId string) string {
	log.Print("Receiving message from queue")
	var result map[string]interface{}
	messages, err := c.receiver.ReceiveMessages(context.TODO(), 100, nil)
	if err != nil {
		panic(err)
	}

	for _, message := range messages {

		body := message.Body
		err := json.Unmarshal(body, &result)

		if result["message_id"] == messageId {
			log.Print("Found message with id: ", messageId)
			err = c.receiver.CompleteMessage(context.TODO(), message, nil)
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
