package cloudsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
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
	Sort      *CloudSearchSortOption
	Highlight []CloudSearchHighlightOption
	// Limit search limit
	Limit int64
	// Offset search offset
	Offset int64
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

type CloudSearchHighlightOption struct {
	Field  string
	Format CloudSearchHighlightFormat
}

type CloudSearchHighlightFormat string

const (
	CloudSearchHighlightFormatHTML = "html"
	CloudSearchHighlightFormatText = "text"
)

// String to highlight query string { "actors": {}, "title": {"format": "text","max_phrases": 2,"pre_tag": "","post_tag": // ""} }
func (h CloudSearchHighlightOption) String() string {
	return fmt.Sprintf("\"%s\": {\"format\": \"%s\"}", h.Field, h.Format)
}

// CloudSearchResponse cloud search response
type CloudSearchResponse struct {
	// Found how many results found
	Found int64
	// Start start from
	Start int64
	// Results search results
	Results []CloudSearchResponseItem
	// Error error
	Error error
}

type CloudSearchResponseItem struct {
	ID         string
	Fields     map[string][]*string
	Highlights map[string]*string
}

func (so CloudSearchSortOption) String() string {
	return fmt.Sprintf("%s %s", so.SortBy, so.Order)
}

type CloudSearchDocumentUploadOptions struct {
	// ContentType content type of document
	ContentType string
	// Content content to be uploaded
	Content []CloudSearchDocumentUploadContent
	// Timeout search timeout
	Timeout time.Duration
}

type CloudSearchDocumentUploadContent struct {
	Fields any                           `json:"fields"`
	Type   CloudSearchDocumentUploadType `json:"type"`
	ID     int64                         `json:"id"`
}

type CloudSearchDocumentUploadType string

const (
	CloudSearchDocumentUploadTypeAdd    CloudSearchDocumentUploadType = "add"
	CloudSearchDocumentUploadTypeDelete CloudSearchDocumentUploadType = "delete"
)

type CloudSearchDocumentUploadResponse struct {
	// Error error
	Error error
	// Adds documents added
	Adds int64
	// Deletes documents deleted
	Deletes int64
}

// Context context includes endpoint, region and bucket info
type context struct {
	searchEndpoint   string
	documentEndpoint string
	region           string
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

// GetSearchEndpoint get search endpoint
func (s *Service) GetSearchEndpoint() string {
	return s.context.searchEndpoint
}

// SetSearchEndpoint set search endpoint
func (s *Service) SetSearchEndpoint(endpoint string) {
	s.context.check()
	s.context.searchEndpoint = endpoint
}

// GetDocumentEndpoint get document endpoint
func (s *Service) GetDocumentEndpoint() string {
	return s.context.documentEndpoint
}

// SetDocumentEndpoint set document endpoint
func (s *Service) SetDocumentEndpoint(endpoint string) {
	s.context.check()
	s.context.documentEndpoint = endpoint
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
func (s *Service) client(search bool) *cloudsearchdomain.CloudSearchDomain {
	endpoint := s.context.searchEndpoint
	if !search {
		endpoint = s.context.documentEndpoint
	}

	once.Do(func() {
		sess, _ := session.NewSession(&aws.Config{
			Region:      aws.String(s.GetRegion()),
			Endpoint:    aws.String(endpoint),
			Credentials: credentials.NewStaticCredentials(s.accessKey, s.accessSecret, ""),
		})

		instance = cloudsearchdomain.New(sess)
	})

	return instance
}

// Search search
func (s *Service) Search(opts *CloudSearchOptions) (resp *CloudSearchResponse) {
	resp = new(CloudSearchResponse)
	client := s.client(true)
	t := 15 * time.Second
	if opts.Timeout > 0 {
		t = opts.Timeout
	}

	if opts.Sort == nil {
		opts.Sort = &DefaultCloudSearchSortOption
	}

	var limit int64 = 100
	if opts.Limit != 0 {
		limit = opts.Limit
	}

	var offset int64 = 0
	if opts.Offset != 0 {
		offset = opts.Offset
	}

	ctx, cancel := goctx.WithTimeout(goctx.Background(), t)
	defer cancel()

	searchinput := &cloudsearchdomain.SearchInput{
		Query: aws.String(opts.Query),
		Sort:  aws.String(opts.Sort.String()),
		Start: aws.Int64(offset),
		Size:  aws.Int64(limit),
	}

	if len(opts.Highlight) > 0 {
		var highlightSb strings.Builder
		highlightSb.WriteString("{")
		for i, opt := range opts.Highlight {
			if i == 0 {
				highlightSb.WriteString(opt.String())
			} else {
				highlightSb.WriteString(fmt.Sprintf(", %s", opt.String()))
			}
		}
		highlightSb.WriteString("}")
		highlight := highlightSb.String()
		searchinput.Highlight = aws.String(highlight)
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

// Upload document upload
func (s *Service) Upload(opts *CloudSearchDocumentUploadOptions) (resp *CloudSearchDocumentUploadResponse) {
	resp = new(CloudSearchDocumentUploadResponse)
	client := s.client(false)

	content, err := json.Marshal(opts.Content)
	if err != nil {
		resp.Error = err
		return
	}

	t := 15 * time.Second
	if opts.Timeout > 0 {
		t = opts.Timeout
	}

	contentType := "application/json"
	if opts.ContentType != "" {
		contentType = opts.ContentType
	}

	ctx, cancel := goctx.WithTimeout(goctx.Background(), t)
	defer cancel()

	uploadinput := &cloudsearchdomain.UploadDocumentsInput{
		ContentType: aws.String(contentType),
		Documents:   bytes.NewReader(content),
	}

	output, err := client.UploadDocumentsWithContext(aws.Context(ctx), uploadinput)
	if err != nil {
		resp.Error = err
		return
	}

	resp.Adds = *output.Adds
	resp.Deletes = *output.Deletes
	return
}

// AsyncUpload async upload
func (s *Service) AsyncUpload(opts *CloudSearchDocumentUploadOptions) (respchan chan<- *CloudSearchDocumentUploadResponse) {
	respchan = make(chan *CloudSearchDocumentUploadResponse)
	go func() {
		respchan <- s.Upload(opts)
	}()
	return respchan
}

func (c *context) check() {
	if c == nil {
		panic("invalid context")
	}
}
