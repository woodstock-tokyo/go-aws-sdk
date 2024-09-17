package agcod

type CreateGiftCardReq struct {
	CreationRequestID string `json:"creationRequestId"`
	PartnerID         string `json:"partnerId"`
	Value             Value  `json:"value"`
}

type Value struct {
	Amount       float64 `json:"amount"`
	CurrencyCode string  `json:"currencyCode"`
}

type CancelGiftCardReq struct {
	CreationRequestID string `json:"creationRequestId"`
	PartnerID         string `json:"partnerId"`
	GcID              string `json:"gcId"`
}
