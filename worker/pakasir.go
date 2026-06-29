package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"mengonten-api/config"
)

type PakasirClient struct {
	Config *config.PakasirConfig
}

type CreateTransactionRequest struct {
	Project string `json:"project"`
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
	APIKey  string `json:"api_key"`
}

type CreateTransactionResponse struct {
	Payment PakasirPayment `json:"payment"`
}

type PakasirPayment struct {
	Project         string `json:"project"`
	OrderID         string `json:"order_id"`
	Amount          int64  `json:"amount"`
	Fee             int64  `json:"fee"`
	TotalPayment    int64  `json:"total_payment"`
	PaymentMethod   string `json:"payment_method"`
	PaymentNumber   string `json:"payment_number"`
	ExpiredAt       string `json:"expired_at"`
}

type CancelTransactionRequest struct {
	Project string `json:"project"`
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
	APIKey  string `json:"api_key"`
}

type WebhookParams struct {
	Amount        int64  `json:"amount"`
	OrderID       string `json:"order_id"`
	Project       string `json:"project"`
	Status        string `json:"status"`
	PaymentMethod string `json:"payment_method"`
	CompletedAt   string `json:"completed_at"`
}

func NewPakasirClient(cfg *config.PakasirConfig) *PakasirClient {
	return &PakasirClient{Config: cfg}
}

var pakasirPaymentMethods = map[string]bool{
	"cimb_niaga_va": true,
	"bni_va":        true,
	"qris":          true,
	"sampoerna_va":  true,
	"bnc_va":        true,
	"maybank_va":    true,
	"permata_va":    true,
	"atm_bersama_va": true,
	"artha_graha_va": true,
	"bri_va":        true,
}

func IsValidPakasirMethod(method string) bool {
	return pakasirPaymentMethods[method]
}

func (c *PakasirClient) CreateTransaction(orderID string, amount int64, method string) (*PakasirPayment, error) {
	if c.Config == nil || c.Config.APIKey == "" {
		return nil, fmt.Errorf("pakasir not configured")
	}

	reqBody := CreateTransactionRequest{
		Project: c.Config.Project,
		OrderID: orderID,
		Amount:  amount,
		APIKey:  c.Config.APIKey,
	}

	bodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/transactioncreate/%s", c.Config.BaseURL, method)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call pakasir: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pakasir returned status %d: %s", resp.StatusCode, string(body))
	}

	var result CreateTransactionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result.Payment, nil
}

func (c *PakasirClient) CancelTransaction(orderID string, amount int64) error {
	if c.Config == nil || c.Config.APIKey == "" {
		return fmt.Errorf("pakasir not configured")
	}

	reqBody := CancelTransactionRequest{
		Project: c.Config.Project,
		OrderID: orderID,
		Amount:  amount,
		APIKey:  c.Config.APIKey,
	}

	bodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.Config.BaseURL+"/api/transactioncancel", bytes.NewReader(bodyJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to call pakasir: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pakasir returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *PakasirClient) VerifyWebhook(params *WebhookParams) bool {
	if c.Config == nil {
		return false
	}
	return params.Project == c.Config.Project && (params.Status == "completed" || params.Status == "success")
}
