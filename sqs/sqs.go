package sqs

import (
	"sync"
	"time"

	goctx "context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// SendMessageOptions send message options
type SendMessageOptions struct {
	Message string
	Timeout time.Duration
}

// SendMessageResponse send message response
type SendMessageResponse struct {
	Error error
}

// ReceiveMessageOptions receive message options
type ReceiveMessageOptions struct {
	// MaxNumberOfMessages how many messages to Receive in one request, max 10
	MaxNumberOfMessages int64
	// WaitTimeSeconds interval for message long polling
	WaitTimeSeconds int64
}

// ReceiveMessageResponse receive message response
type ReceiveMessageResponse struct {
	Messages []*ReceiveMessage
	Error    error
}

// DeleteMessageOptions delete message options
type DeleteMessageOptions struct {
	ReceiptHandle string
	Timeout       time.Duration
}

// ReceiveMessageResponse delete message response
type DeleteMessageResponse struct {
	Error error
}

// GetQueueAttributesOptions get queue attributes options
type GetQueueAttributesOptions struct {
	AttributeNames []*string
	Timeout        time.Duration
}

// GetQueueAttributesResponse get queue attributes response
type GetQueueAttributesResponse struct {
	Error      error
	Attributes map[string]*string
}

type ReceiveMessage struct {
	Message       string
	ReceiptHandle string
}

// Context context includes endpoint, region and bucket info
type context struct {
	region string
	queue  string
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

// SetQueue set quene
func (s *Service) SetQueue(queue string) {
	s.context.check()
	s.context.queue = queue
}

// GetQueue get queue
func (s *Service) GetQueue() string {
	return s.context.queue
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

var once sync.Once
var instance *sqs.SQS

// client init client
func (s *Service) client() *sqs.SQS {
	once.Do(func() {
		sess, _ := session.NewSession(&aws.Config{
			Region:      aws.String(s.GetRegion()),
			Credentials: credentials.NewStaticCredentials(s.accessKey, s.accessSecret, ""),
		})

		instance = sqs.New(sess)
	})

	return instance
}

// SendMessage send message
func (s *Service) SendMessage(opts *SendMessageOptions) (resp *SendMessageResponse) {
	resp = new(SendMessageResponse)

	client := s.client()
	t := 30 * time.Second
	if opts.Timeout > 0 {
		t = opts.Timeout
	}
	ctx, cancel := goctx.WithTimeout(goctx.Background(), t)
	defer cancel()

	_, err := client.SendMessageWithContext(ctx, &sqs.SendMessageInput{
		MessageBody: aws.String(opts.Message),
		QueueUrl:    aws.String(s.GetQueue()),
	})

	if err != nil {
		resp.Error = err
	}

	return
}

// AsyncSendMessage async send message
func (s *Service) AsyncSendMessage(opts *SendMessageOptions) (respchan chan<- *SendMessageResponse) {
	respchan = make(chan *SendMessageResponse)
	go func() {
		respchan <- s.SendMessage(opts)
	}()
	return respchan
}

// ReceiveMessage receive message
func (s *Service) ReceiveMessage(opts *ReceiveMessageOptions) (resp *ReceiveMessageResponse) {
	resp = new(ReceiveMessageResponse)

	client := s.client()
	if opts.MaxNumberOfMessages == 0 || opts.MaxNumberOfMessages > 10 {
		opts.MaxNumberOfMessages = 1
	}

	if opts.WaitTimeSeconds == 0 {
		opts.WaitTimeSeconds = 10
	}

	sqsResp, err := client.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.GetQueue()),
		MaxNumberOfMessages: aws.Int64(opts.MaxNumberOfMessages),
		WaitTimeSeconds:     aws.Int64(opts.WaitTimeSeconds),
	})

	if err != nil {
		resp.Error = err
	} else {
		messages := []*ReceiveMessage{}
		for _, message := range sqsResp.Messages {
			messages = append(messages, &ReceiveMessage{
				Message:       *message.Body,
				ReceiptHandle: *message.ReceiptHandle,
			})
		}
		resp.Messages = messages
	}

	return
}

// AsyncReceiveMessage async receive message
func (s *Service) AsyncReceiveMessage(opts *ReceiveMessageOptions) (respchan chan<- *ReceiveMessageResponse) {
	respchan = make(chan *ReceiveMessageResponse)
	go func() {
		respchan <- s.ReceiveMessage(opts)
	}()
	return respchan
}

// DeleteMessage delete message
func (s *Service) DeleteMessage(opts *DeleteMessageOptions) (resp *DeleteMessageResponse) {
	resp = new(DeleteMessageResponse)

	client := s.client()
	t := 30 * time.Second
	if opts.Timeout > 0 {
		t = opts.Timeout
	}
	ctx, cancel := goctx.WithTimeout(goctx.Background(), t)
	defer cancel()

	_, err := client.DeleteMessageWithContext(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.GetQueue()),
		ReceiptHandle: aws.String(opts.ReceiptHandle),
	})

	if err != nil {
		resp.Error = err
	}

	return
}

// AsyncDeleteMessage async delete message
func (s *Service) AsyncDeleteMessage(opts *DeleteMessageOptions) (respchan chan<- *DeleteMessageResponse) {
	respchan = make(chan *DeleteMessageResponse)
	go func() {
		respchan <- s.DeleteMessage(opts)
	}()
	return respchan
}

func (c *context) check() {
	if c == nil {
		panic("invalid context")
	}
}

// CountMessage count messages in a queue
func (s *Service) GetQueueAttributes(opts *GetQueueAttributesOptions) (resp *GetQueueAttributesResponse) {
	resp = new(GetQueueAttributesResponse)

	client := s.client()
	t := 30 * time.Second
	if opts.Timeout > 0 {
		t = opts.Timeout
	}
	ctx, cancel := goctx.WithTimeout(goctx.Background(), t)
	defer cancel()

	attributes, err := client.GetQueueAttributesWithContext(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(s.GetQueue()),
		AttributeNames: opts.AttributeNames,
	})

	if err != nil {
		resp.Error = err
	}

	resp.Attributes = attributes.Attributes
	return
}
