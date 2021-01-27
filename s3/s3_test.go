package s3

import (
	"testing"
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
	if resp.Error != nil {
		t.Error(resp.Error)
	}
}
