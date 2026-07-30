package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type UserServiceClient interface {
	ValidateUser(ctx context.Context, userID uuid.UUID) (bool, error)
}

type userServiceClient struct {
	baseURL string
	http    *http.Client
}

func NewUserServiceClient(baseURL string) UserServiceClient {
	return &userServiceClient{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *userServiceClient) ValidateUser(ctx context.Context, userID uuid.UUID) (bool, error) {
	url := fmt.Sprintf("%s/internal/users/%s", c.baseURL, userID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("user service returned status %d", resp.StatusCode)
	}

	var result struct {
		Exists bool `json:"exists"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}
	return result.Exists, nil
}
