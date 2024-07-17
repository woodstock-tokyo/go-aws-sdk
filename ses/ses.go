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

// CreateTemplateOptions create template options
type CreateTemplateOptions struct {
	TemplateName string
	Subject      string
	HTML         *string
	Text         *string
}

// CreateTemplateResponse create template response
type CreateTemplateResponse struct {
	Error error
}

// UpdateTemplateOptions update template options
type UpdateTemplateOptions struct {
	TemplateName string
	Subject      string
	HTML         *string
	Text         *string
}

// UpdateTemplateResponse create template response
type UpdateTemplateResponse struct {
	Error error
}

// DeleteTemplateOptions delete template options
type DeleteTemplateOptions struct {
	TemplateName string
}

// DeleteTemplateResponse delete template response
type DeleteTemplateResponse struct {
	Error error
}

// ListTemplatesOptions list templates options
type ListTemplatesOptions struct {
	// MaxItems max is 10
	MaxItems  *int64
	NextToken *string
}

// ListTemplatesResponse list templates response
type ListTemplatesResponse struct {
	Error     error
	Templates []string
}

// GetTemplateOptions get templates options
type GetTemplateOptions struct {
	TemplateName string
}

// GetTemplateResponse get templates response
type GetTemplateResponse struct {
	Error        error
	SubjectPart  *string
	TemplateName *string
	TextPart     *string
	HtmlPart     *string
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

// SendEmail send email
func (s *Service) SendEmail(opts *SendEmailOptions) (resp *SendEmailResponse) {
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
		respchan <- s.SendEmail(opts)
	}()
	return respchan
}

func (c *context) check() {
	if c == nil {
		panic("invalid context")
	}
}

// CreateTemplate creates an SES email template
func (s *Service) CreateTemplate(opts *CreateTemplateOptions) (resp *CreateTemplateResponse) {
	resp = new(CreateTemplateResponse)
	client := s.client()
	input := &ses.CreateTemplateInput{
		Template: &ses.Template{
			TemplateName: aws.String(opts.TemplateName),
			SubjectPart:  aws.String(opts.Subject),
			HtmlPart:     opts.HTML,
			TextPart:     opts.Text,
		},
	}

	_, err := client.CreateTemplate(input)
	if err != nil {
		resp.Error = err
	}

	return
}

// UpdateTemplate updates an existing SES email template
func (s *Service) UpdateTemplate(opts *UpdateTemplateOptions) (resp *UpdateTemplateResponse) {
	resp = new(UpdateTemplateResponse)
	client := s.client()
	input := &ses.UpdateTemplateInput{
		Template: &ses.Template{
			TemplateName: aws.String(opts.TemplateName),
			SubjectPart:  aws.String(opts.Subject),
			HtmlPart:     opts.HTML,
			TextPart:     opts.Text,
		},
	}

	_, err := client.UpdateTemplate(input)
	if err != nil {
		resp.Error = err
	}

	return
}

// DeleteTemplate deletes an SES email template by name
func (s *Service) DeleteTemplate(opts *DeleteTemplateOptions) (resp *DeleteTemplateResponse) {
	resp = new(DeleteTemplateResponse)
	client := s.client()
	input := &ses.DeleteTemplateInput{
		TemplateName: aws.String(opts.TemplateName),
	}

	_, err := client.DeleteTemplate(input)
	if err != nil {
		resp.Error = err
	}

	return
}

// ListTemplates lists SES email templates
func (s *Service) ListTemplates(opts *ListTemplatesOptions) (resp *ListTemplatesResponse) {
	resp = new(ListTemplatesResponse)
	client := s.client()
	input := &ses.ListTemplatesInput{
		MaxItems:  opts.MaxItems,
		NextToken: opts.NextToken,
	}

	result, err := client.ListTemplates(input)
	if err != nil {
		resp.Error = err
	}

	for _, template := range result.TemplatesMetadata {
		resp.Templates = append(resp.Templates, *template.Name)
	}

	return
}

// GetTemplate retrieves details of an SES email template by name
func (s *Service) GetTemplate(opts *GetTemplateOptions) (resp *GetTemplateResponse) {
	resp = new(GetTemplateResponse)
	client := s.client()
	input := &ses.GetTemplateInput{
		TemplateName: aws.String(opts.TemplateName),
	}

	result, err := client.GetTemplate(input)
	if err != nil {
		resp.Error = err
	}

	resp.SubjectPart = result.Template.SubjectPart
	resp.TemplateName = result.Template.TemplateName
	resp.TextPart = result.Template.TextPart
	resp.HtmlPart = result.Template.HtmlPart

	return
}
