package secretsmanager

import (
	"encoding/json"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// GetSecretResponse response for get secret
type GetSecretResponse struct {
	SecretValue map[string]string
	Error       error
}

// Context context includes region
type context struct {
	region string
}

// Service service includes context and credentials
type Service struct {
	context *context
}

// NewService service initializer
func NewService() *Service {
	return &Service{
		context: new(context),
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
var instance *secretsmanager.SecretsManager

// client init client
func (s *Service) client() *secretsmanager.SecretsManager {
	once.Do(func() {
		sess, _ := session.NewSession(&aws.Config{
			Region: aws.String(s.GetRegion()),
		})
		instance = secretsmanager.New(sess)
	})

	return instance
}

// GetSecretValue get secret value by secret id, secret id can be either ARN or secret name
func (s *Service) GetSecretValue(secretID string) (resp *GetSecretResponse) {
	resp = new(GetSecretResponse)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	}

	result, err := s.client().GetSecretValue(input)
	if err != nil {
		resp.Error = err
		return
	}

	secretvalue := map[string]string{}
	err = json.Unmarshal([]byte(aws.StringValue(result.SecretString)), &secretvalue)
	resp.SecretValue = secretvalue
	resp.Error = err
	return
}

// AsyncGetSecretValue async get secret value
func (s *Service) AsyncGetSecretValue(secretID string) (respchan chan<- *GetSecretResponse) {
	respchan = make(chan *GetSecretResponse)
	go func() {
		respchan <- s.GetSecretValue(secretID)
	}()
	return respchan
}

func (c *context) check() {
	if c == nil {
		panic("invalid context")
	}
}
