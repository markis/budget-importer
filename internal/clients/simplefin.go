package clients

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/markis/budget-importer/internal/models"
)

// SimpleFinClient handles communication with the SimpleFIN API.
type SimpleFinClient struct {
	httpClient *http.Client
	accessURL  string
	username   string
	password   string
}

// NewSimpleFinClient creates a new SimpleFIN client.
func NewSimpleFinClient(accessURL, username, password string) *SimpleFinClient {
	return &SimpleFinClient{
		accessURL:  accessURL,
		username:   username,
		password:   password,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchTransactions fetches transactions from SimpleFIN starting from the given date.
func (c *SimpleFinClient) FetchTransactions(startDate time.Time) (*models.SimpleFinResponse, error) {
	// Build the URL with query parameters.
	baseURL := c.accessURL + "/accounts"
	params := url.Values{}
	params.Set("pending", "1")
	params.Set("start-date", strconv.FormatInt(startDate.Unix(), 10))

	fullURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequest(http.MethodGet, fullURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set Basic Auth header.
	auth := base64.StdEncoding.EncodeToString([]byte(c.username + ":" + c.password))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("SimpleFIN API returned status %d", resp.StatusCode)
		}

		return nil, fmt.Errorf("SimpleFIN API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response models.SimpleFinResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}
