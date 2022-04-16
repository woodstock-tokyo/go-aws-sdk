package dynamo

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

// Context _context includes endpoint, region
type _context struct {
	endpoint string
	region   string
}

// Service service includes context and credentials
type Service struct {
	context      *_context
	accessKey    string
	accessSecret string
}

// NewService service initializer
func NewService(key, secret string) *Service {
	return &Service{
		context:      new(_context),
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

// SetRegion set endpoint
func (s *Service) SetEndpoint(endpoint string) {
	s.context.check()
	s.context.endpoint = endpoint
}

// GetRegion get endpoint
func (s *Service) GetEndpoint() string {
	return s.context.endpoint
}

var once sync.Once
var instance *DB

// Instance init DB instance
func (s *Service) Instance() *DB {
	once.Do(func() {
		sess := session.New(&aws.Config{
			Region:      aws.String(s.GetRegion()),
			Endpoint:    aws.String(s.GetEndpoint()),
			Credentials: credentials.NewStaticCredentials(s.accessKey, s.accessSecret, ""),
		})

		instance = New(sess)
	})

	return instance
}

func (c *_context) check() {
	if c == nil {
		panic("invalid context")
	}
}
