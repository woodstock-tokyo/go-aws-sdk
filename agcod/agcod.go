package agcod

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
)

// Service service includes context and credentials
type Service struct {
	context      *context
	accessKey    string
	accessSecret string
	baseURL      string
}

// Context context includes endpoint, region and other info
type context struct {
	region    string
	token     string
	service   string
	partnerID string
}

func (c *context) check() {
	if c == nil {
		panic("invalid context")
	}
}

// NewService service initializer
// See reference https://developer.amazon.com/ja/docs/incentives-api/gift-codes-on-demand.html#creategiftcard
func NewService(key, secret, baseURL string) *Service {
	return &Service{
		context:      new(context),
		accessKey:    key,
		accessSecret: secret,
		baseURL:      baseURL,
	}
}

// SetRegion set region
func (s *Service) SetRegion(region string) {
	s.context.check()
	s.context.region = region
}

// SetToken set token
func (s *Service) SetToken(token string) {
	s.context.check()
	s.context.token = token
}

// SetService set token
func (s *Service) SetService(service string) {
	s.context.check()
	s.context.service = service
}

// SetPartnerID partner id
func (s *Service) SetPartnerID(partnerID string) {
	s.context.check()
	s.context.partnerID = partnerID
}

// SetSignatureV4 set signature v4
func (s *Service) SetSignatureV4(req *http.Request, signTime time.Time) error {
	creds := credentials.NewStaticCredentials(s.accessKey, s.accessSecret, s.context.token)
	signer := v4.NewSigner(creds)

	if req.Body != nil {
		defer req.Body.Close()
		data, err := io.ReadAll(req.Body)
		if err != nil {
			return err
		}
		_, err = signer.Sign(req, bytes.NewReader(data), s.context.service, s.context.region, signTime)
		if err != nil {
			return err
		}
	} else {
		_, err := signer.Sign(req, nil, s.context.service, s.context.region, signTime)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateGiftCard /CreateGiftCard
func (s *Service) CreateGiftCard(getGiftCardReq CreateGiftCardReq, signTime time.Time) (*CreateGiftCardResp, error) {
	url := fmt.Sprintf("%s/CreateGiftCard", s.baseURL)
	_getGiftCardReq, err := json.Marshal(getGiftCardReq)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(_getGiftCardReq))
	req.Header.Set("x-amz-target", "com.amazonaws.agcod.AGCODService.CreateGiftCard")
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")

	err = s.SetSignatureV4(req, signTime)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var createGiftCardResp *CreateGiftCardResp
	if err := json.Unmarshal(body, &createGiftCardResp); err != nil {
		return nil, err
	}

	return createGiftCardResp, nil
}

// CancelGiftCard /CancelGiftCard
func (s *Service) CancelGiftCard(cancelGiftCardReq CancelGiftCardReq, signTime time.Time) (*CancelGiftCardResp, error) {
	url := fmt.Sprintf("%s/CancelGiftCard", s.baseURL)
	_cancelGiftCardReq, err := json.Marshal(cancelGiftCardReq)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(_cancelGiftCardReq))
	req.Header.Set("x-amz-target", "com.amazonaws.agcod.AGCODService.CancelGiftCard")
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")

	err = s.SetSignatureV4(req, signTime)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var cancelGiftCardResp *CancelGiftCardResp
	if err := json.Unmarshal(body, &cancelGiftCardResp); err != nil {
		return nil, err
	}

	return cancelGiftCardResp, nil
}

// GetAvailableFunds /GetAvailableFunds
func (s *Service) GetAvailableFunds(partnerID string, signTime time.Time) (*GetAvailableFundsResp, error) {
	url := fmt.Sprintf("%s/GetAvailableFunds", s.baseURL)
	getAvailableFundsReq := fmt.Sprintf("{ \"partnerId\": \"%s\" }", partnerID)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(getAvailableFundsReq)))
	req.Header.Set("x-amz-target", "com.amazonaws.agcod.AGCODService.GetAvailableFunds")
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")

	err := s.SetSignatureV4(req, signTime)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var getAvailableFundsResp *GetAvailableFundsResp
	if err := json.Unmarshal(body, &getAvailableFundsResp); err != nil {
		return nil, err
	}

	return getAvailableFundsResp, nil
}
