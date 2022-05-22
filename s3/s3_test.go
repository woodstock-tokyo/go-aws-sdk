package s3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUpload test object upload
func TestUpload(t *testing.T) {
	svc := NewService("<<access key>>", "<<secret key>>")
	svc.SetRegion("<<region>>")
	svc.SetBucket("<<bucket name>>")

	opts := &UploadOptions{
		FileName:       "./test.png",
		Public:         true,
		Encryption:     KMS,
		EncrptionKeyID: "<<KMS key id>>",
	}

	resp := svc.Upload(opts)
	assert.NoError(t, resp.Error)
}

// TestList test object list
func TestList(t *testing.T) {
	svc := NewService("<<access key>>", "<<secret key>>")
	svc.SetRegion("<<region>>")
	svc.SetBucket("<<bucket name>>")

	opts := &ListOptions{
		Prefix: "picture/",
	}

	resp := svc.List(opts)
	assert.NoError(t, resp.Error)
}
