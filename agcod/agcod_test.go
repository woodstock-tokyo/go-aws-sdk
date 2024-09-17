package agcod

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	sandboxAccessKey = ""
	sandboxSecretKey = ""
	sandboxURL       = "https://agcod-v2-fe-gamma.amazon.com"
)

func TestGETSignature(t *testing.T) {
	accessKey := "AKID"
	secretkey := "SECRET"
	token := "SESSION"
	region := "us-east-1"
	service := "es"
	signTime := time.Unix(0, 0)

	expect := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s/%s/%s/aws4_request, SignedHeaders=host;x-amz-date;x-amz-security-token, Signature=6601e883cc6d23871fd6c2a394c5677ea2b8c82b04a6446786d64cd74f520967", accessKey, signTime.Format("20060102"), region, service)

	url := "https://subdomain.us-east-1.es.amazonaws.com/log-*/_search"
	req, _ := http.NewRequest("GET", url, nil)

	s := NewService(accessKey, secretkey, "")
	s.SetRegion(region)
	s.SetService(service)
	s.SetToken(token)
	err := s.SetSignatureV4(req, signTime)
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}

	assert.Equal(t, expect, req.Header.Get("Authorization"))
}

// you can create expected signature here https://s3.amazonaws.com/AGCOD/htmlSDKv2/htmlSDKv2_NAEUFE/index.html
func TestPOSTSignature(t *testing.T) {
	accessKey := "AKID"
	secretkey := "SECRET"
	token := ""
	region := "us-west-2"
	service := "AGCODService"
	signTime := time.Unix(1467963107, 0) // 20160708T073147Z

	expect := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s/%s/%s/aws4_request, SignedHeaders=accept;host;x-amz-date;x-amz-target, Signature=a47dd06c1bcff61e8c96fee2c87c4230a4dd5f7d4ffb5b958897027a3acae54b", accessKey, signTime.Format("20060102"), region, service)

	jsonStr := `{"creationRequestId":"","partnerId":"Amazon","value":{"currencyCode":"USD","amount":null}}`
	url := "https://agcod-v2-fe-gamma.amazon.com/CreateGiftCard"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonStr)))
	req.Header.Set("x-amz-target", "com.amazonaws.agcod.AGCODService.CreateGiftCard")
	req.Header.Set("accept", "application/json")

	s := NewService(accessKey, secretkey, "")
	s.SetRegion(region)
	s.SetService(service)
	s.SetToken(token)
	err := s.SetSignatureV4(req, signTime)
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}

	assert.Equal(t, expect, req.Header.Get("Authorization"))
}

func TestCreateGiftCard(t *testing.T) {
	s := NewService(sandboxAccessKey, sandboxSecretKey, sandboxURL)
	s.SetRegion("us-west-2")
	s.SetService("AGCODService")

	resp, err := s.CreateGiftCard(CreateGiftCardReq{
		CreationRequestID: "Wo0f5123",
		PartnerID:         "Wo0f5",
		Value: Value{
			Amount:       5.00,
			CurrencyCode: "JPY",
		},
	}, time.Now())

	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}

	assert.NotNil(t, resp)
}

func TestCancelGiftCard(t *testing.T) {
	s := NewService(sandboxAccessKey, sandboxSecretKey, sandboxURL)
	s.SetRegion("us-west-2")
	s.SetService("AGCODService")

	resp, err := s.CancelGiftCard(CancelGiftCardReq{
		CreationRequestID: "Wo0f5123",
		PartnerID:         "Wo0f5",
		GcID:              "5512629721495365",
	}, time.Now())

	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}

	assert.NotNil(t, resp)
}

func TestGetAvailableFunds(t *testing.T) {
	s := NewService(sandboxAccessKey, sandboxSecretKey, sandboxURL)
	s.SetRegion("us-west-2")
	s.SetService("AGCODService")

	resp, err := s.GetAvailableFunds("Wo0f5", time.Now())

	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}

	assert.NotNil(t, resp)
}
