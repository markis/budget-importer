package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// SimpleFinOrganization represents the organization associated with an account.
type SimpleFinOrganization struct {
	Domain  string  `json:"domain"`
	Name    string  `json:"name"`
	SfinURL *string `json:"sfin-url,omitempty"`
}

// SimpleFinHolding represents an investment holding.
type SimpleFinHolding struct {
	CostBasis     string `json:"cost_basis"`
	Currency      string `json:"currency"`
	Description   string `json:"description"`
	ID            string `json:"id"`
	MarketValue   string `json:"market_value"`
	PurchasePrice string `json:"purchase_price"`
	Shares        string `json:"shares"`
	Symbol        string `json:"symbol"`
	Created       int64  `json:"created"`
}

// SimpleFinTransaction represents a financial transaction.
type SimpleFinTransaction struct {
	ID           string          `json:"id"`
	Description  string          `json:"description"`
	Memo         string          `json:"memo"`
	Payee        string          `json:"payee"`
	Amount       decimal.Decimal `json:"amount"`
	Posted       int64           `json:"posted"`
	TransactedAt int64           `json:"transacted_at"`

	// Fields added during processing.
	Category *string `json:"-"`
}

// PostedTime returns the posted timestamp as time.Time.
func (t *SimpleFinTransaction) PostedTime() time.Time {
	return time.Unix(t.Posted, 0)
}

// TransactedAtTime returns the transacted_at timestamp as time.Time.
func (t *SimpleFinTransaction) TransactedAtTime() time.Time {
	return time.Unix(t.TransactedAt, 0)
}

// SimpleFinAccount represents a financial account.
type SimpleFinAccount struct {
	AvailableBalance string                 `json:"available-balance"`
	Balance          string                 `json:"balance"`
	Currency         string                 `json:"currency"`
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Org              SimpleFinOrganization  `json:"org"`
	Holdings         []SimpleFinHolding     `json:"holdings"`
	Transactions     []SimpleFinTransaction `json:"transactions"`
	BalanceDate      int64                  `json:"balance-date"`
}

// SimpleFinResponse represents the API response from SimpleFIN.
type SimpleFinResponse struct {
	Accounts    []SimpleFinAccount `json:"accounts"`
	Errors      []string           `json:"errors,omitempty"`
	XAPIMessage []string           `json:"x-api-message,omitempty"`
}
