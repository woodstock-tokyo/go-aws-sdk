package sns

import (
	"sync"
	"time"

	goctx "context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

// AddEndpointOptions options to add a mobile push endpoint
type AddEndpointOptions struct {
	PlatformApplicationArn string
	Token                  string
	Attributes             map[string]string
	Timeout                time.Duration
}

// AddEndpointResponse response for adding a mobile push endpoint
type AddEndpointResponse struct {
	EndpointArn string
	Error       error
}

// DeleteEndpointOptions options to delete a mobile push endpoint
type DeleteEndpointOptions struct {
	EndpointArn string
	Timeout     time.Duration
}

// DeleteEndpointResponse response for deleting a mobile push endpoint
type DeleteEndpointResponse struct {
	Error error
}

// SubscribeOptions options to subscribe to a topic
type SubscribeOptions struct {
	TopicArn string
	Protocol string
	Endpoint string
	Timeout  time.Duration
}

// SubscribeResponse response for subscribing to a topic
type SubscribeResponse struct {
	SubscriptionArn string
	Error           error
}

// UnsubscribeOptions options to unsubscribe from a topic
type UnsubscribeOptions struct {
	SubscriptionArn string
	Timeout         time.Duration
}

// UnsubscribeResponse response for unsubscribing from a topic
type UnsubscribeResponse struct {
	Error error
}

// ListSubscribersOptions options to list subscribers
type ListSubscribersOptions struct {
	TopicArn string
	Timeout  time.Duration
}

// ListSubscribersResponse response for listing subscribers
type ListSubscribersResponse struct {
	Subscribers []*sns.Subscription
	Error       error
}

// Context context includes endpoint, region and other info
type context struct {
	region   string
	topicArn string
}

// Service service includes context and credentials
type Service struct {
	context      *context
	accessKey    string
	accessSecret string
}

// NewService service initializer
func NewService(key, secret string) *Service {
	return &Service{
		context:      new(context),
		accessKey:    key,
		accessSecret: secret,
	}
}

// SetRegion set region
func (s *Service) SetRegion(region string) {
	s.context.check()
	s.context.region = region
}

// GetRegion get region
func (s *Service) GetRegion() string {
	return s.context.region
}

// SetTopicArn set topic arn
func (s *Service) SetTopicArn(topicArn string) {
	s.context.check()
	s.context.topicArn = topicArn
}

// GetTopicArn get topic arn
func (s *Service) GetTopicArn() string {
	return s.context.topicArn
}

var once sync.Once
var instance *sns.SNS

// client init client
func (s *Service) client() *sns.SNS {
	once.Do(func() {
		sess, _ := session.NewSession(&aws.Config{
			Region:      aws.String(s.GetRegion()),
			Credentials: credentials.NewStaticCredentials(s.accessKey, s.accessSecret, ""),
		})

		instance = sns.New(sess)
	})

	return instance
}

// AddEndpoint adds a mobile push endpoint to SNS
func (s *Service) AddEndpoint(opts *AddEndpointOptions) (resp *AddEndpointResponse) {
	resp = new(AddEndpointResponse)

	client := s.client()
	t := 30 * time.Second
	if opts.Timeout > 0 {
		t = opts.Timeout
	}
	ctx, cancel := goctx.WithTimeout(goctx.Background(), t)
	defer cancel()

	input := &sns.CreatePlatformEndpointInput{
		PlatformApplicationArn: aws.String(opts.PlatformApplicationArn),
		Token:                  aws.String(opts.Token),
		Attributes:             aws.StringMap(opts.Attributes),
	}

	output, err := client.CreatePlatformEndpointWithContext(ctx, input)
	if err != nil {
		resp.Error = err
		return
	}

	resp.EndpointArn = *output.EndpointArn
	return
}

// DeleteEndpoint removes a mobile push endpoint from SNS
func (s *Service) DeleteEndpoint(opts *DeleteEndpointOptions) (resp *DeleteEndpointResponse) {
	resp = new(DeleteEndpointResponse)

	client := s.client()
	t := 30 * time.Second
	if opts.Timeout > 0 {
		t = opts.Timeout
	}
	ctx, cancel := goctx.WithTimeout(goctx.Background(), t)
	defer cancel()

	input := &sns.DeleteEndpointInput{
		EndpointArn: aws.String(opts.EndpointArn),
	}

	_, err := client.DeleteEndpointWithContext(ctx, input)
	if err != nil {
		resp.Error = err
	}

	return
}

// Subscribe subscribes an endpoint to an SNS topic
func (s *Service) Subscribe(opts *SubscribeOptions) (resp *SubscribeResponse) {
	resp = new(SubscribeResponse)

	client := s.client()
	t := 30 * time.Second
	if opts.Timeout > 0 {
		t = opts.Timeout
	}
	ctx, cancel := goctx.WithTimeout(goctx.Background(), t)
	defer cancel()

	input := &sns.SubscribeInput{
		TopicArn: aws.String(opts.TopicArn),
		Protocol: aws.String(opts.Protocol),
		Endpoint: aws.String(opts.Endpoint),
	}

	output, err := client.SubscribeWithContext(ctx, input)
	if err != nil {
		resp.Error = err
		return
	}

	resp.SubscriptionArn = *output.SubscriptionArn
	return
}

// Unsubscribe unsubscribes an endpoint from an SNS topic
func (s *Service) Unsubscribe(opts *UnsubscribeOptions) (resp *UnsubscribeResponse) {
	resp = new(UnsubscribeResponse)

	client := s.client()
	t := 30 * time.Second
	if opts.Timeout > 0 {
		t = opts.Timeout
	}
	ctx, cancel := goctx.WithTimeout(goctx.Background(), t)
	defer cancel()

	input := &sns.UnsubscribeInput{
		SubscriptionArn: aws.String(opts.SubscriptionArn),
	}

	_, err := client.UnsubscribeWithContext(ctx, input)
	if err != nil {
		resp.Error = err
	}

	return
}

// ListSubscribers lists all subscribers for a given SNS topic
func (s *Service) ListSubscribers(opts *ListSubscribersOptions) (resp *ListSubscribersResponse) {
	resp = new(ListSubscribersResponse)

	client := s.client()
	t := 30 * time.Second
	if opts.Timeout > 0 {
		t = opts.Timeout
	}
	ctx, cancel := goctx.WithTimeout(goctx.Background(), t)
	defer cancel()

	var subscriptions []*sns.Subscription
	err := client.ListSubscriptionsByTopicPagesWithContext(ctx, &sns.ListSubscriptionsByTopicInput{
		TopicArn: aws.String(opts.TopicArn),
	}, func(page *sns.ListSubscriptionsByTopicOutput, lastPage bool) bool {
		subscriptions = append(subscriptions, page.Subscriptions...)
		return !lastPage
	})

	if err != nil {
		resp.Error = err
		return
	}

	resp.Subscribers = subscriptions
	return
}

func (c *context) check() {
	if c == nil {
		panic("invalid context")
	}
}
