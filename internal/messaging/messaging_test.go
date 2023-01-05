package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"reflect"
	"testing"
)

type mockSender struct{}

func (s *mockSender) SendMessage(ctx context.Context, message *azservicebus.Message, opts *azservicebus.SendMessageOptions) error {
	return nil
}

func (s *mockSender) Close(ctx context.Context) error {
	return nil
}

type mockReceiver struct {
	messages []Message
}

func (r *mockReceiver) ReceiveMessages(ctx context.Context, maxMessages int, options *azservicebus.ReceiveMessagesOptions) ([]*azservicebus.ReceivedMessage, error) {
	var sbMessages []*azservicebus.ReceivedMessage
	for _, m := range r.messages {
		body, _ := json.Marshal(m)
		sbMessage := &azservicebus.ReceivedMessage{
			Body: body,
		}
		sbMessages = append(sbMessages, sbMessage)
	}
	return sbMessages, nil
}

func (r *mockReceiver) CompleteMessage(ctx context.Context, message *azservicebus.ReceivedMessage, opts *azservicebus.CompleteMessageOptions) error {
	return nil
}

func (r *mockReceiver) Close(ctx context.Context) error {
	return nil
}

func TestQueueClient_SendMessage(t *testing.T) {
	queueClient := &QueueClient{
		sender: &mockSender{},
	}

	testCases := []struct {
		name        string
		message     Message
		expectedErr error
	}{
		{
			name: "Success: Send message with valid data",
			message: Message{
				MessageId: "123",
				Uuid:      "456",
				Prompt:    "test prompt",
				Strength:  0.5,
			},
			expectedErr: nil,
		},
		{
			name:        "Error: Send message with empty data",
			message:     Message{},
			expectedErr: fmt.Errorf("json: error calling Marshal: empty struct"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := queueClient.SendMessage(tc.message)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Errorf("Expected error: %v, but got: %v", tc.expectedErr, err)
			}
		})
	}
}

func TestQueueClient_ReceiveMessage(t *testing.T) {
	queueClient := &QueueClient{
		receiver: &mockReceiver{},
	}

	testCases := []struct {
		name             string
		messageId        string
		expectedResult   string
		expectedErr      error
		existingMessages []Message
	}{
		{
			name:           "Success: Receive message with valid ID",
			messageId:      "123",
			expectedResult: "456",
			expectedErr:    nil,
			existingMessages: []Message{
				{
					MessageId: "123",
					Uuid:      "456",
				},
			},
		},
		{
			name:           "Error: Receive message with invalid ID",
			messageId:      "123",
			expectedResult: "",
			expectedErr:    nil,
			existingMessages: []Message{
				{
					MessageId: "456",
					Uuid:      "789",
				},
			},
		},
		{
			name:             "Success: Receive message when there are no messages",
			messageId:        "123",
			expectedResult:   "",
			expectedErr:      nil,
			existingMessages: []Message{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Mock the receiver to return existing messages
			queueClient.receiver.(*mockReceiver).messages = tc.existingMessages

			result := queueClient.ReceiveMessage(tc.messageId)
			if result != tc.expectedResult {
				t.Errorf("Expected result: %s, but got: %s", tc.expectedResult, result)
			}

			err := queueClient.receiver.Close(context.Background())
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Errorf("Expected error: %v, but got: %v", tc.expectedErr, err)
			}
		})
	}
}
