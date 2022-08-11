package cloudsearch

import (
	"fmt"
	"sync"
	"time"

	goctx "context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudsearchdomain"
)

// CloudSearchOptions cloud search options
type CloudSearchOptions struct {
	// query to search
	Query string
	// Sort by
	Sort *CloudSearchSortOption
	// Limit search limit
	Limit *int64
	// Offset search offset
	Offset *int64
	// Timeout search timeout
	Timeout time.Duration
}

// https://docs.aws.amazon.com/cloudsearch/latest/developerguide/sorting-results.html
type CloudSearchSortOrder string

const (
	CloudSearchSortOrderAsc  = "asc"
	CloudSearchSortOrderDesc = "desc"
)

type CloudSearchSortOption struct {
	SortBy string
	Order  CloudSearchSortOrder
}

var DefaultCloudSearchSortOption = CloudSearchSortOption{
	SortBy: "_score",
	Order:  CloudSearchSortOrderDesc,
}

// CloudSearchResponse cloud search response
type CloudSearchResponse struct {
	// Found how many results found
	Found int64
	// Start start from
	Start int64
	// results
	Results []CloudSearchResponseItem
	Error   error
}

type CloudSearchResponseItem struct {
	ID         string
	Fields     map[string][]*string
	Highlights map[string]*string
}

func (so CloudSearchSortOption) String() string {
	return fmt.Sprintf("%s %s", so.SortBy, so.Order)
}

// Context context includes endpoint, region and bucket info
type context struct {
	endpoint string
	region   string
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

// GetEndpoint get endpoint
func (s *Service) GetEndpoint() string {
	return s.context.endpoint
}

// SetEndpoint set endpoint
func (s *Service) SetEndpoint(endpoint string) {
	s.context.check()
	s.context.endpoint = endpoint
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
var instance *cloudsearchdomain.CloudSearchDomain

// client init client
func (s *Service) client() *cloudsearchdomain.CloudSearchDomain {
	once.Do(func() {
		sess, _ := session.NewSession(&aws.Config{
			Region:      aws.String(s.GetRegion()),
			Endpoint:    aws.String(s.context.endpoint),
			Credentials: credentials.NewStaticCredentials(s.accessKey, s.accessSecret, ""),
		})

		instance = cloudsearchdomain.New(sess)
	})

	return instance
}

// Search search
func (s *Service) Search(opts *CloudSearchOptions) (resp *CloudSearchResponse) {
	resp = new(CloudSearchResponse)
	client := s.client()
	t := 15 * time.Second
	if opts.Timeout > 0 {
		t = opts.Timeout
	}

	if opts.Sort == nil {
		opts.Sort = &DefaultCloudSearchSortOption
	}

	var limit int64 = 100
	if opts.Limit != nil {
		limit = *opts.Limit
	}

	var offset int64 = 0
	if opts.Offset != nil {
		offset = *opts.Offset
	}

	ctx, cancel := goctx.WithTimeout(goctx.Background(), t)
	defer cancel()

	searchinput := &cloudsearchdomain.SearchInput{
		Query: aws.String(opts.Query),
		Sort:  aws.String(opts.Sort.String()),
		Start: aws.Int64(offset),
		Size:  aws.Int64(limit),
	}

	output, err := client.SearchWithContext(aws.Context(ctx), searchinput)
	if err != nil {
		resp.Error = err
		return
	}

	resp.Found = *output.Hits.Found
	resp.Start = *output.Hits.Start

	resp.Results = []CloudSearchResponseItem{}
	for _, hit := range output.Hits.Hit {
		resp.Results = append(resp.Results, CloudSearchResponseItem{
			ID:         *hit.Id,
			Fields:     hit.Fields,
			Highlights: hit.Highlights,
		})
	}
	return
}

// AsyncSearch async search
func (s *Service) AsyncSearch(opts *CloudSearchOptions) (respchan chan<- *CloudSearchResponse) {
	respchan = make(chan *CloudSearchResponse)
	go func() {
		respchan <- s.Search(opts)
	}()
	return respchan
}

func (c *context) check() {
	if c == nil {
		panic("invalid context")
	}
}
