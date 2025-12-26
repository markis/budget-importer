package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/markis/budget-importer/internal/models"
)

func TestSimpleFinTransaction_PostedTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		posted   int64
		expected time.Time
	}{
		{
			name:     "zero timestamp",
			posted:   0,
			expected: time.Unix(0, 0),
		},
		{
			name:     "specific timestamp",
			posted:   1703548800, // 2023-12-26 00:00:00 UTC
			expected: time.Unix(1703548800, 0),
		},
		{
			name:     "negative timestamp",
			posted:   -86400,
			expected: time.Unix(-86400, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			txn := &models.SimpleFinTransaction{Posted: tt.posted}
			result := txn.PostedTime()

			if !result.Equal(tt.expected) {
				t.Errorf("PostedTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSimpleFinTransaction_TransactedAtTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		transactedAt int64
		expected     time.Time
	}{
		{
			name:         "zero timestamp",
			transactedAt: 0,
			expected:     time.Unix(0, 0),
		},
		{
			name:         "specific timestamp",
			transactedAt: 1703548800,
			expected:     time.Unix(1703548800, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			txn := &models.SimpleFinTransaction{TransactedAt: tt.transactedAt}
			result := txn.TransactedAtTime()

			if !result.Equal(tt.expected) {
				t.Errorf("TransactedAtTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSimpleFinResponse_JSONUnmarshal(t *testing.T) {
	t.Parallel()

	jsonData := `{
		"accounts": [
			{
				"id": "acc123",
				"name": "Checking Account",
				"balance": "1000.50",
				"available-balance": "950.00",
				"balance-date": 1703548800,
				"currency": "USD",
				"org": {
					"domain": "bank.com",
					"name": "Test Bank"
				},
				"transactions": [
					{
						"id": "txn456",
						"amount": -25.99,
						"description": "Coffee Shop",
						"memo": "Morning coffee",
						"payee": "Starbucks",
						"posted": 1703548800,
						"transacted_at": 1703545200
					}
				],
				"holdings": []
			}
		],
		"errors": ["warning message"]
	}`

	var response models.SimpleFinResponse

	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(response.Accounts) != 1 {
		t.Errorf("Expected 1 account, got %d", len(response.Accounts))
	}

	account := response.Accounts[0]

	if account.ID != "acc123" {
		t.Errorf("Expected account ID 'acc123', got '%s'", account.ID)
	}

	if account.Name != "Checking Account" {
		t.Errorf("Expected account name 'Checking Account', got '%s'", account.Name)
	}

	if account.Balance != "1000.50" {
		t.Errorf("Expected balance '1000.50', got '%s'", account.Balance)
	}

	if account.Org.Name != "Test Bank" {
		t.Errorf("Expected org name 'Test Bank', got '%s'", account.Org.Name)
	}

	if len(account.Transactions) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(account.Transactions))
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

	if len(response.Errors) != 1 || response.Errors[0] != "warning message" {
		t.Errorf("Expected errors ['warning message'], got %v", response.Errors)
	}
}

func TestSimpleFinOrganization_JSONUnmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		json     string
		expected models.SimpleFinOrganization
	}{
		{
			name: "with sfin-url",
			json: `{"domain": "bank.com", "name": "Test Bank", "sfin-url": "https://sfin.bank.com"}`,
			expected: models.SimpleFinOrganization{
				Domain:  "bank.com",
				Name:    "Test Bank",
				SfinURL: strPtr("https://sfin.bank.com"),
			},
		},
		{
			name: "without sfin-url",
			json: `{"domain": "bank.com", "name": "Test Bank"}`,
			expected: models.SimpleFinOrganization{
				Domain:  "bank.com",
				Name:    "Test Bank",
				SfinURL: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var org models.SimpleFinOrganization

			err := json.Unmarshal([]byte(tt.json), &org)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			if org.Domain != tt.expected.Domain {
				t.Errorf("Expected domain '%s', got '%s'", tt.expected.Domain, org.Domain)
			}

			if org.Name != tt.expected.Name {
				t.Errorf("Expected name '%s', got '%s'", tt.expected.Name, org.Name)
			}

			if tt.expected.SfinURL == nil {
				if org.SfinURL != nil {
					t.Errorf("Expected nil SfinURL, got '%s'", *org.SfinURL)
				}
			} else {
				if org.SfinURL == nil || *org.SfinURL != *tt.expected.SfinURL {
					t.Errorf("Expected SfinURL '%s', got %v", *tt.expected.SfinURL, org.SfinURL)
				}
			}
		})
	}
}

func TestSimpleFinHolding_JSONUnmarshal(t *testing.T) {
	t.Parallel()

	jsonData := `{
		"id": "hold123",
		"cost_basis": "1000.00",
		"currency": "USD",
		"description": "Tech Stock",
		"market_value": "1250.00",
		"purchase_price": "100.00",
		"shares": "10",
		"symbol": "TECH",
		"created": 1703548800
	}`

	var holding models.SimpleFinHolding

	err := json.Unmarshal([]byte(jsonData), &holding)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if holding.ID != "hold123" {
		t.Errorf("Expected ID 'hold123', got '%s'", holding.ID)
	}

	if holding.Symbol != "TECH" {
		t.Errorf("Expected symbol 'TECH', got '%s'", holding.Symbol)
	}

	if holding.Shares != "10" {
		t.Errorf("Expected shares '10', got '%s'", holding.Shares)
	}

	if holding.Created != 1703548800 {
		t.Errorf("Expected created 1703548800, got %d", holding.Created)
	}
}

func strPtr(s string) *string {
	return &s
}
