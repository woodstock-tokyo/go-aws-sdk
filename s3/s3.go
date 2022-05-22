package s3

import (
	"bytes"
	goctx "context"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// EncrptionType Encryption Type
type EncrptionType string

const (
	// None do not encrypt
	None EncrptionType = ""
	// AES256 AES256
	AES256 EncrptionType = "AES256"
	// KMS KMS
	KMS EncrptionType = "aws:kms"
)

// UploadOptions upload options
type UploadOptions struct {
	// filename to upload
	FileName string
	// assign bucket subdirectory, otherwise object will be saved under root
	SubDirectory string
	// visible to public or not
	Public bool
	// Timeout upload timeout
	Timeout time.Duration
	// Encryption encryption type
	Encryption EncrptionType
	// EncrptionKeyID Encrption key ID (only supports "aws:kms" type)
	EncrptionKeyID string
	// Attachment download (true) or show inline when open s3 uri in browser
	Attachment bool
}

// UploadResponse upload response
type UploadResponse struct {
	// Location location of uploaded file
	Location string
	Error    error
}

// ListOptions list obj options
type ListOptions struct {
	// assign prefix to list, otherwise object will be saved under root
	Prefix string
	// Timeout upload timeout
	Timeout time.Duration
}

// ListResponse list response
type ListResponse struct {
	Objects []ListObject
	Error   error
}

// ListObject list object
type ListObject struct {
	Key          string
	LastModified time.Time
	Size         int64
}

// Context context includes endpoint, region and bucket info
type context struct {
	endpoint string
	region   string
	bucket   string
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

// GetEndPoint get endpoint
func (s *Service) GetEndPoint() string {
	return "s3.amazonaws.com"
}

// SetBucket set bucket
func (s *Service) SetBucket(bucket string) {
	s.context.check()
	s.context.bucket = bucket
}

// GetBucket get bucket
func (s *Service) GetBucket() string {
	return s.context.bucket
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
var instance *s3.S3

// client init client
func (s *Service) client() *s3.S3 {
	once.Do(func() {
		sess := session.New(&aws.Config{
			Region:      aws.String(s.GetRegion()),
			Credentials: credentials.NewStaticCredentials(s.accessKey, s.accessSecret, ""),
		})

		instance = s3.New(sess)
	})

	return instance
}

// Upload upload file
func (s *Service) Upload(opts *UploadOptions) (resp *UploadResponse) {
	resp = new(UploadResponse)

	client := s.client()
	t := 180 * time.Second
	if opts.Timeout > 0 {
		t = opts.Timeout
	}
	ctx, cancel := goctx.WithTimeout(goctx.Background(), t)
	defer cancel()

	filenamme := opts.FileName
	objname := resolveObjName(opts.SubDirectory, filenamme)
	contenttype, err := resolveContentType(filenamme)
	if err != nil {
		contenttype = ""
	}

	file, err := os.Open(filenamme)
	if err != nil {
		resp.Error = fmt.Errorf("failed to read local file")
	}
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	putobjectinput := &s3.PutObjectInput{
		Bucket:        aws.String(s.GetBucket()),
		Key:           aws.String(objname),
		Body:          bytes.NewReader(buffer),
		ContentLength: aws.Int64(size),
	}

	if contenttype != "" {
		putobjectinput.ContentType = aws.String(contenttype)
	}

	if opts.Attachment {
		putobjectinput.ContentDisposition = aws.String("attachment")
	}

	if opts.Public {
		putobjectinput.ACL = aws.String("public-read")
	} else {
		putobjectinput.ACL = aws.String("private")
	}

	if opts.Encryption == AES256 {
		putobjectinput.ServerSideEncryption = aws.String(string(AES256))
	} else if opts.Encryption == KMS {
		putobjectinput.ServerSideEncryption = aws.String(string(KMS))
		if opts.EncrptionKeyID != "" {
			putobjectinput.SSEKMSKeyId = aws.String(opts.EncrptionKeyID)
		}
	}

	_, err = client.PutObjectWithContext(aws.Context(ctx), putobjectinput)
	if err != nil {
		resp.Error = err
	} else {
		resp.Location = fmt.Sprintf("https://s3-%s.amazonaws.com/%s/%s", s.GetRegion(), s.GetBucket(), objname)
	}

	return
}

// AsyncUpload async upload
func (s *Service) AsyncUpload(opts *UploadOptions) (respchan chan<- *UploadResponse) {
	respchan = make(chan *UploadResponse)
	go func() {
		respchan <- s.Upload(opts)
	}()
	return respchan
}

// List list files
func (s *Service) List(opts *ListOptions) (resp *ListResponse) {
	resp = &ListResponse{
		Objects: []ListObject{},
	}

	client := s.client()
	t := 10 * time.Second
	if opts.Timeout > 0 {
		t = opts.Timeout
	}
	ctx, cancel := goctx.WithTimeout(goctx.Background(), t)
	defer cancel()

	listobjinput := &s3.ListObjectsInput{
		Bucket: aws.String(s.GetBucket()),
		Prefix: aws.String(opts.Prefix),
	}

	list, err := client.ListObjectsWithContext(ctx, listobjinput)
	if err != nil {
		resp.Error = err
	} else {
		for _, obj := range list.Contents {
			resp.Objects = append(resp.Objects, ListObject{
				Key:          aws.StringValue(obj.Key),
				LastModified: aws.TimeValue(obj.LastModified),
				Size:         aws.Int64Value(obj.Size),
			})
		}
	}

	return
}

func resolveObjName(subdiresctory string, fullfilename string) string {
	// doesn't matter if subdirectory is empty string
	return path.Join(subdiresctory, filepath.Base(fullfilename))
}

func resolveContentType(fullfilename string) (string, error) {
	f, err := os.Open(fullfilename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)
	_, err = f.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)
	return contentType, nil
}

func (c *context) check() {
	if c == nil {
		panic("invalid context")
	}
}
