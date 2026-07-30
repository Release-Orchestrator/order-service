package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// PaymentResult contains the result from processing a payment.
type PaymentResult struct {
	PaymentID uuid.UUID `json:"paymentId"`
	Status    string    `json:"status"`
}

// PaymentServiceClient defines operations for communicating with the payment service.
type PaymentServiceClient interface {
	ProcessPayment(ctx context.Context, orderID uuid.UUID, amount float64) (*PaymentResult, error)
}

type paymentServiceClient struct {
	baseURL string
	http    *http.Client
}

// NewPaymentServiceClient creates a new PaymentServiceClient.
func NewPaymentServiceClient(baseURL string) PaymentServiceClient {
	return &paymentServiceClient{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// ProcessPayment sends a payment request to the payment service.
func (c *paymentServiceClient) ProcessPayment(ctx context.Context, orderID uuid.UUID, amount float64) (*PaymentResult, error) {
	url := fmt.Sprintf("%s/internal/payments", c.baseURL)

	body := map[string]interface{}{
		"orderId": orderID.String(),
		"amount":  amount,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("warning: failed to close response body: %v", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("payment service returned status %d", resp.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			PaymentID uuid.UUID `json:"paymentId"`
			Status    string    `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("payment service returned unsuccessful response")
	}

	return &PaymentResult{
		PaymentID: result.Data.PaymentID,
		Status:    result.Data.Status,
	}, nil
}
