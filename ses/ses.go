package ses

import (
	"encoding/json"
	"sync"
	"time"

	goctx "context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

// SendEmailOptions send email options
type SendEmailOptions struct {
	Sender           string
	Recipients       []string
	CCs              []string
	BCCs             []string
	Template         string
	TemplateData     map[string]string
	ConfigurationSet *string
	Timeout          time.Duration
}

// SendEmailResponse send email response
type SendEmailResponse struct {
	Error error
}

// Context context includes endpoint, region and other info
type context struct {
	region string
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

var once sync.Once
var instance *ses.SES

// client init client
func (s *Service) client() *ses.SES {
	once.Do(func() {
		sess, _ := session.NewSession(&aws.Config{
			Region:      aws.String(s.GetRegion()),
			Credentials: credentials.NewStaticCredentials(s.accessKey, s.accessSecret, ""),
		})

		instance = ses.New(sess)
	})

	return instance
}

// SendTampleteEmail send email
func (s *Service) SendTampleteEmail(opts *SendEmailOptions) (resp *SendEmailResponse) {
	resp = new(SendEmailResponse)

	client := s.client()
	t := 30 * time.Second
	if opts.Timeout > 0 {
		t = opts.Timeout
	}
	ctx, cancel := goctx.WithTimeout(goctx.Background(), t)
	defer cancel()

	var toAddresses []*string
	for _, recipient := range opts.Recipients {
		toAddresses = append(toAddresses, aws.String(recipient))
	}

	var ccs []*string
	for _, cc := range opts.CCs {
		ccs = append(ccs, aws.String(cc))
	}

	var bccs []*string
	for _, bcc := range opts.BCCs {
		bccs = append(bccs, aws.String(bcc))
	}

	input := &ses.SendTemplatedEmailInput{
		Source: aws.String(opts.Sender),
		Destination: &ses.Destination{
			ToAddresses:  toAddresses,
			CcAddresses:  ccs,
			BccAddresses: bccs,
		},
		Template:             aws.String(opts.Template),
		ConfigurationSetName: opts.ConfigurationSet,
	}

	templateDataJson, err := json.Marshal(opts.TemplateData)
	if err != nil {
		resp.Error = err
		return
	}

	input.TemplateData = aws.String(string(templateDataJson))

	_, err = client.SendTemplatedEmailWithContext(ctx, input)
	if err != nil {
		resp.Error = err
	}

	return
}

// AsyncSendTamplateEmail async send email
func (s *Service) AsyncSendTamplateEmail(opts *SendEmailOptions) (respchan chan<- *SendEmailResponse) {
	respchan = make(chan *SendEmailResponse)
	go func() {
		respchan <- s.SendTampleteEmail(opts)
	}()
	return respchan
}

func (c *context) check() {
	if c == nil {
		panic("invalid context")
	}
}
