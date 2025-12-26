package clients_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/markis/budget-importer/internal/clients"
)

func TestNewSimpleFinClient(t *testing.T) {
	t.Parallel()

	client := clients.NewSimpleFinClient("https://api.simplefin.org", "user", "pass")

	if client == nil {
		t.Fatal("Expected non-nil client")
	}
}

func TestSimpleFinClient_FetchTransactions_Success(t *testing.T) {
	t.Parallel()

	response := map[string]any{
		"accounts": []map[string]any{
			{
				"id":                "acc123",
				"name":              "Checking",
				"balance":           "1000.00",
				"available-balance": "950.00",
				"balance-date":      1703548800,
				"currency":          "USD",
				"org": map[string]string{
					"domain": "bank.com",
					"name":   "Test Bank",
				},
				"transactions": []map[string]any{
					{
						"id":            "txn456",
						"amount":        -25.99,
						"description":   "Coffee Shop",
						"memo":          "Morning coffee",
						"payee":         "Starbucks",
						"posted":        1703548800,
						"transacted_at": 1703545200,
					},
					{
						"id":            "txn789",
						"amount":        100.00,
						"description":   "Deposit",
						"memo":          "",
						"payee":         "Direct Deposit",
						"posted":        1703548800,
						"transacted_at": 1703548800,
					},
				},
				"holdings": []any{},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request path.
		if r.URL.Path != "/accounts" {
			t.Errorf("Expected path '/accounts', got '%s'", r.URL.Path)
		}

		// Verify query parameters.
		if r.URL.Query().Get("pending") != "1" {
			t.Errorf("Expected pending=1, got '%s'", r.URL.Query().Get("pending"))
		}

		if r.URL.Query().Get("start-date") == "" {
			t.Error("Expected start-date parameter")
		}

		// Verify authorization header.
		auth := r.Header.Get("Authorization")
		if auth == "" {
			t.Error("Expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := clients.NewSimpleFinClient(server.URL, "user", "pass")
	startDate := time.Now().AddDate(0, 0, -2)

	result, err := client.FetchTransactions(startDate)
	if err != nil {
		t.Fatalf("FetchTransactions failed: %v", err)
	}

	if len(result.Accounts) != 1 {
		t.Fatalf("Expected 1 account, got %d", len(result.Accounts))
	}

	account := result.Accounts[0]

	if account.ID != "acc123" {
		t.Errorf("Expected account ID 'acc123', got '%s'", account.ID)
	}

	if account.Name != "Checking" {
		t.Errorf("Expected account name 'Checking', got '%s'", account.Name)
	}

	if len(account.Transactions) != 2 {
		t.Fatalf("Expected 2 transactions, got %d", len(account.Transactions))
	}

	txn := account.Transactions[0]

	if txn.ID != "txn456" {
		t.Errorf("Expected transaction ID 'txn456', got '%s'", txn.ID)
	}

	expectedAmount := decimal.NewFromFloat(-25.99)
	if !txn.Amount.Equal(expectedAmount) {
		t.Errorf("Expected amount %v, got %v", expectedAmount, txn.Amount)
	}

	if txn.Payee != "Starbucks" {
		t.Errorf("Expected payee 'Starbucks', got '%s'", txn.Payee)
	}
}

func TestSimpleFinClient_FetchTransactions_HTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := clients.NewSimpleFinClient(server.URL, "user", "pass")
	startDate := time.Now().AddDate(0, 0, -2)

	_, err := client.FetchTransactions(startDate)
	if err == nil {
		t.Fatal("Expected error for HTTP 500 response")
	}
}

func TestSimpleFinClient_FetchTransactions_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := clients.NewSimpleFinClient(server.URL, "user", "pass")
	startDate := time.Now().AddDate(0, 0, -2)

	_, err := client.FetchTransactions(startDate)
	if err == nil {
		t.Fatal("Expected error for invalid JSON response")
	}
}

func TestSimpleFinClient_FetchTransactions_EmptyResponse(t *testing.T) {
	t.Parallel()

	response := map[string]any{
		"accounts": []any{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := clients.NewSimpleFinClient(server.URL, "user", "pass")
	startDate := time.Now().AddDate(0, 0, -2)

	result, err := client.FetchTransactions(startDate)
	if err != nil {
		t.Fatalf("FetchTransactions failed: %v", err)
	}

	if len(result.Accounts) != 0 {
		t.Errorf("Expected 0 accounts, got %d", len(result.Accounts))
	}
}

func TestSimpleFinClient_FetchTransactions_WithErrors(t *testing.T) {
	t.Parallel()

	response := map[string]any{
		"accounts": []any{},
		"errors":   []string{"Bank connection timeout", "Partial data"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := clients.NewSimpleFinClient(server.URL, "user", "pass")
	startDate := time.Now().AddDate(0, 0, -2)

	result, err := client.FetchTransactions(startDate)
	if err != nil {
		t.Fatalf("FetchTransactions failed: %v", err)
	}

	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(result.Errors))
	}

	if result.Errors[0] != "Bank connection timeout" {
		t.Errorf("Expected first error 'Bank connection timeout', got '%s'", result.Errors[0])
	}
}

func TestSimpleFinClient_FetchTransactions_Unauthorized(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	client := clients.NewSimpleFinClient(server.URL, "wrong", "creds")
	startDate := time.Now().AddDate(0, 0, -2)

	_, err := client.FetchTransactions(startDate)
	if err == nil {
		t.Fatal("Expected error for unauthorized response")
	}
}

func TestSimpleFinClient_FetchTransactions_ConnectionError(t *testing.T) {
	t.Parallel()

	// Use an invalid URL to simulate connection error.
	client := clients.NewSimpleFinClient("http://localhost:99999", "user", "pass")
	startDate := time.Now().AddDate(0, 0, -2)

	_, err := client.FetchTransactions(startDate)
	if err == nil {
		t.Fatal("Expected error for connection failure")
	}
}

func TestSimpleFinClient_FetchTransactions_InvalidURL(t *testing.T) {
	t.Parallel()

	// Use an invalid URL that will fail at request creation.
	client := clients.NewSimpleFinClient("://invalid", "user", "pass")
	startDate := time.Now().AddDate(0, 0, -2)

	_, err := client.FetchTransactions(startDate)
	if err == nil {
		t.Fatal("Expected error for invalid URL")
	}
}

func TestSimpleFinClient_FetchTransactions_HTTPErrorWithoutBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Set content length but don't write body to simulate read error.
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusBadRequest)
		// Don't write the full body, causing a potential read error.
	}))
	defer server.Close()

	client := clients.NewSimpleFinClient(server.URL, "user", "pass")
	startDate := time.Now().AddDate(0, 0, -2)

	_, err := client.FetchTransactions(startDate)
	if err == nil {
		t.Fatal("Expected error for HTTP error response")
	}
}

func TestSimpleFinClient_FetchTransactions_VerifyAuthHeader(t *testing.T) {
	t.Parallel()

	var capturedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"accounts":[]}`))
	}))
	defer server.Close()

	client := clients.NewSimpleFinClient(server.URL, "testuser", "testpass")
	startDate := time.Now().AddDate(0, 0, -2)

	_, err := client.FetchTransactions(startDate)
	if err != nil {
		t.Fatalf("FetchTransactions failed: %v", err)
	}

	// Verify the auth header contains "Basic " prefix.
	if capturedAuth == "" {
		t.Error("Expected Authorization header to be set")
	}

	if len(capturedAuth) < 7 || capturedAuth[:6] != "Basic " {
		t.Errorf("Expected Authorization header to start with 'Basic ', got '%s'", capturedAuth)
	}
}
