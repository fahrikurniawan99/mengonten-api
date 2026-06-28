package worker

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"mengonten-api/config"
)

type DuitkuClient struct {
	Config *config.DuitkuConfig
}

type CreateInvoiceRequest struct {
	PaymentAmount   int64  `json:"paymentAmount"`
	MerchantOrderID string `json:"merchantOrderId"`
	ProductDetails  string `json:"productDetails"`
	Email           string `json:"email"`
	CallbackURL     string `json:"callbackUrl"`
	ReturnURL       string `json:"returnUrl"`
}

type CreateInvoiceResponse struct {
	MerchantCode string `json:"merchantCode"`
	Reference    string `json:"reference"`
	PaymentURL   string `json:"paymentUrl"`
}

type CallbackParams struct {
	MerchantCode    string `form:"merchantCode"`
	Amount          string `form:"amount"`
	MerchantOrderID string `form:"merchantOrderId"`
	ProductDetail   string `form:"productDetail"`
	AdditionalParam string `form:"additionalParam"`
	PaymentCode     string `form:"paymentCode"`
	ResultCode      string `form:"resultCode"`
	MerchantUserID  string `form:"merchantUserId"`
	Reference       string `form:"reference"`
	Signature       string `form:"signature"`
	PublisherOrderID string `form:"publisherOrderId"`
}

func NewDuitkuClient(cfg *config.DuitkuConfig) *DuitkuClient {
	return &DuitkuClient{Config: cfg}
}

func (d *DuitkuClient) CreateInvoice(amount int64, orderID, email string) (*CreateInvoiceResponse, error) {
	if d.Config == nil || d.Config.MerchantCode == "" {
		return nil, fmt.Errorf("duitku not configured")
	}

	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())

	stringToSign := d.Config.MerchantCode + timestamp
	signature := computeHMACSHA256(stringToSign, d.Config.APIKey)

	reqBody := CreateInvoiceRequest{
		PaymentAmount:   amount,
		MerchantOrderID: orderID,
		ProductDetails:  "Langganan Mengonten",
		Email:           email,
		CallbackURL:     d.Config.CallbackURL,
		ReturnURL:       d.Config.ReturnURL,
	}

	bodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", d.Config.BaseURL+"/api/merchant/createInvoice", bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-duitku-timestamp", timestamp)
	httpReq.Header.Set("x-duitku-signature", signature)
	httpReq.Header.Set("x-duitku-merchantcode", d.Config.MerchantCode)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call duitku: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("duitku returned status %d: %s", resp.StatusCode, string(body))
	}

	var result CreateInvoiceResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

func (d *DuitkuClient) VerifyCallback(params *CallbackParams) bool {
	if d.Config == nil || d.Config.MerchantKey == "" {
		return false
	}

	stringToSign := params.MerchantCode + params.Amount + params.MerchantOrderID
	expectedSig := computeHMACSHA256(stringToSign, d.Config.MerchantKey)

	return hmac.Equal([]byte(expectedSig), []byte(params.Signature))
}

func computeHMACSHA256(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
