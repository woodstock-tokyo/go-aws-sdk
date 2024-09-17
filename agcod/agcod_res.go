package agcod

type CreateGiftCardResp struct {
	GcClaimCode       string `json:"gcClaimCode"`
	CardInfo          CardInfo
	GcID              string `json:"gcId"`
	CreationRequestID string `json:"creationRequestId"`
	GcExpirationDate  string `json:"gcExpirationDate"`
	Status            string `json:"status"`
}

type CardInfo struct {
	CardNumber     string `json:"cardNumber"`
	CardStatus     string `json:"cardStatus"`
	ExpirationDate string `json:"expirationDate"`
	Value          Value
}

type CancelGiftCardResp struct {
	CreationRequestID string `json:"creationRequestId"`
	GcID              string `json:"gcId"`
	Status            string `json:"status"`
}

type GetAvailableFundsResp struct {
	AvailableFunds Value
	Status         string `json:"status"`
	Timestamp      string `json:"timestamp"`
}
