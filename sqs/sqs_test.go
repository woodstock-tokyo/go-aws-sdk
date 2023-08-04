package sqs

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSendMessage test send message
func TestSendMessage(t *testing.T) {
	svc := NewService(os.Getenv("WS_SQS_AWS_ACCESS_KEY_ID"), os.Getenv("WS_SQS_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")
	svc.SetQueue("push-notification-stg")

	opts := &SendMessageOptions{
		Message: "whoisyourdaddy",
	}

	resp := svc.SendMessage(opts)
	assert.NoError(t, resp.Error)
}

// TestReceiveMessage test receive message
func TestReceiveMessage(t *testing.T) {
	svc := NewService(os.Getenv("WS_SQS_AWS_ACCESS_KEY_ID"), os.Getenv("WS_SQS_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")
	svc.SetQueue("push-notification-stg")

	receiveOpts := &ReceiveMessageOptions{
		MaxNumberOfMessages: 1,
	}

	// ensure message is sent
	time.Sleep(3 * time.Second)

	receiveResp := svc.ReceiveMessage(receiveOpts)
	assert.NoError(t, receiveResp.Error)

	message := receiveResp.Messages[0]
	assert.Equal(t, "whoisyourdaddy", message.Message)

	// delete it
	deleteOpts := &DeleteMessageOptions{
		ReceiptHandle: message.ReceiptHandle,
	}

	deleteResp := svc.DeleteMessage(deleteOpts)
	assert.NoError(t, deleteResp.Error)
}
