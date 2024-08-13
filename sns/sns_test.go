package sns

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddEndpoint(t *testing.T) {
	svc := NewService(os.Getenv("WS_SNS_AWS_ACCESS_KEY_ID"), os.Getenv("WS_SNS_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")
	opts := &AddEndpointOptions{
		PlatformApplicationArn: "arn:aws:sns:ap-northeast-1:324792451081:app/GCM/push-notification-stg",
		Token:                  "test-device-token",
		Attributes:             map[string]string{"CustomUserData": "user123"},
	}
	resp := svc.AddEndpoint(opts)
	assert.NoError(t, resp.Error)
	assert.NotEmpty(t, resp.EndpointArn)
}

func TestSubscribe(t *testing.T) {
	svc := NewService(os.Getenv("WS_SNS_AWS_ACCESS_KEY_ID"), os.Getenv("WS_SNS_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")
	endpointArn := "arn:aws:sns:ap-northeast-1:324792451081:endpoint/GCM/push-notification-stg/63832370-4b8a-30ee-8d03-39e59a94a01a"
	opts := &SubscribeOptions{
		TopicArn: "arn:aws:sns:ap-northeast-1:324792451081:push-notifications-all-users-stg",
		Protocol: "application",
		Endpoint: endpointArn,
	}
	resp := svc.Subscribe(opts)
	assert.NoError(t, resp.Error)
	assert.NotEmpty(t, resp.SubscriptionArn)
}

func TestListSubscribers(t *testing.T) {
	svc := NewService(os.Getenv("WS_SNS_AWS_ACCESS_KEY_ID"), os.Getenv("WS_SNS_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")

	opts := &ListSubscribersOptions{
		TopicArn: "arn:aws:sns:ap-northeast-1:324792451081:push-notifications-all-users-stg",
		Timeout:  10 * time.Second,
	}

	resp := svc.ListSubscribers(opts)
	assert.NoError(t, resp.Error)
	assert.NotNil(t, resp.Subscribers)
	assert.GreaterOrEqual(t, len(resp.Subscribers), 1, "Expected at least one subscriber")

	for _, sub := range resp.Subscribers {
		assert.NotEmpty(t, *sub.SubscriptionArn)
		assert.NotEmpty(t, *sub.Protocol)
		assert.NotEmpty(t, *sub.Endpoint)
	}
}

func TestUnsubscribe(t *testing.T) {
	svc := NewService(os.Getenv("WS_SNS_AWS_ACCESS_KEY_ID"), os.Getenv("WS_SNS_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")
	subscriptionArn := "arn:aws:sns:ap-northeast-1:324792451081:push-notifications-all-users-stg:bb206965-f3d6-49df-b643-2a07abf236a1"
	opts := &UnsubscribeOptions{
		SubscriptionArn: subscriptionArn,
	}
	resp := svc.Unsubscribe(opts)
	assert.NoError(t, resp.Error)
}

func TestDeleteEndpoint(t *testing.T) {
	svc := NewService(os.Getenv("WS_SNS_AWS_ACCESS_KEY_ID"), os.Getenv("WS_SNS_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")
	endpointArn := "arn:aws:sns:ap-northeast-1:324792451081:endpoint/GCM/push-notification-stg/63832370-4b8a-30ee-8d03-39e59a94a01a"
	opts := &DeleteEndpointOptions{
		EndpointArn: endpointArn,
	}
	resp := svc.DeleteEndpoint(opts)
	assert.NoError(t, resp.Error)
}
