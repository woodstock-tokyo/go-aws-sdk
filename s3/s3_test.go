package s3

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUpload test object upload
func TestUpload(t *testing.T) {
	svc := NewService(os.Getenv("WS_S3_AWS_ACCESS_KEY_ID"), os.Getenv("WS_S3_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")
	svc.SetBucket("woodstock-static-hosting")

	opts := &UploadOptions{
		FileName: "./test.png",
		Public:   true,
		// Encryption:     KMS,
		// EncrptionKeyID: "<<KMS key id>>",
	}

	resp := svc.Upload(opts)
	assert.NoError(t, resp.Error)

	opts.FileName = "./test.txt"
	resp = svc.Upload(opts)
	assert.NoError(t, resp.Error)
}

// TestList test object list
func TestList(t *testing.T) {
	svc := NewService(os.Getenv("WS_S3_AWS_ACCESS_KEY_ID"), os.Getenv("WS_S3_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")
	svc.SetBucket("woodstock-static-hosting")

	opts := &ListOptions{
		Prefix: "picture/",
	}

	resp := svc.List(opts)
	assert.NoError(t, resp.Error)
}
