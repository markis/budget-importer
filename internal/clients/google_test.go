package clients

import (
	"testing"

	"github.com/markis/budget-importer/internal/models"
)

func TestParseRow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		row              []any
		expectedCategory *string
		expectedName     *string
		expectNil        bool
	}{
		{
			name:      "empty row",
			row:       []any{},
			expectNil: true,
		},
		{
			name:             "payee only",
			row:              []any{"Starbucks"},
			expectedCategory: nil,
			expectedName:     nil,
			expectNil:        false,
		},
		{
			name:             "payee and category",
			row:              []any{"Starbucks", "Food"},
			expectedCategory: strPtr("Food"),
			expectedName:     nil,
			expectNil:        false,
		},
		{
			name:             "payee, category, and name",
			row:              []any{"Starbucks", "Food", "Coffee Shop"},
			expectedCategory: strPtr("Food"),
			expectedName:     strPtr("Coffee Shop"),
			expectNil:        false,
		},
		{
			name:             "empty category string",
			row:              []any{"Starbucks", "", "Coffee Shop"},
			expectedCategory: nil,
			expectedName:     strPtr("Coffee Shop"),
			expectNil:        false,
		},
		{
			name:             "empty name string",
			row:              []any{"Starbucks", "Food", ""},
			expectedCategory: strPtr("Food"),
			expectedName:     nil,
			expectNil:        false,
		},
		{
			name:             "non-string category",
			row:              []any{"Starbucks", 123, "Coffee Shop"},
			expectedCategory: nil,
			expectedName:     strPtr("Coffee Shop"),
			expectNil:        false,
		},
		{
			name:             "non-string name",
			row:              []any{"Starbucks", "Food", 456},
			expectedCategory: strPtr("Food"),
			expectedName:     nil,
			expectNil:        false,
		},
		{
			name:             "all empty strings",
			row:              []any{"Starbucks", "", ""},
			expectedCategory: nil,
			expectedName:     nil,
			expectNil:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := parseRow(tt.row)

			if tt.expectNil {
				if result != nil {
					t.Errorf("Expected nil result, got %+v", result)
				}

				return
			}

			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			if tt.expectedCategory == nil {
				if result.Category != nil {
					t.Errorf("Expected nil category, got '%s'", *result.Category)
				}
			} else {
				if result.Category == nil {
					t.Errorf("Expected category '%s', got nil", *tt.expectedCategory)
				} else if *result.Category != *tt.expectedCategory {
					t.Errorf("Expected category '%s', got '%s'", *tt.expectedCategory, *result.Category)
				}
			}

			if tt.expectedName == nil {
				if result.Name != nil {
					t.Errorf("Expected nil name, got '%s'", *result.Name)
				}
			} else {
				if result.Name == nil {
					t.Errorf("Expected name '%s', got nil", *tt.expectedName)
				} else if *result.Name != *tt.expectedName {
					t.Errorf("Expected name '%s', got '%s'", *tt.expectedName, *result.Name)
				}
			}
		})
	}
}

func TestGoogleSheetRow_Conversion(t *testing.T) {
	t.Parallel()

	// Test that GoogleSheetRow can be converted to [][]any for the API.
	rows := []models.GoogleSheetRow{
		{"txn1", "Payee1", 10.50, "1/1/2024", "Food", ""},
		{"txn2", "Payee2", -25.00, "1/2/2024", "Transport", "http://receipt.url"},
	}

	values := make([][]any, len(rows))
	for i, row := range rows {
		values[i] = row
	}

	if len(values) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(values))
	}

	if len(values[0]) != 6 {
		t.Errorf("Expected 6 columns in row 0, got %d", len(values[0]))
	}

	if values[0][0] != "txn1" {
		t.Errorf("Expected values[0][0] = 'txn1', got %v", values[0][0])
	}

	if values[1][2] != -25.00 {
		t.Errorf("Expected values[1][2] = -25.00, got %v", values[1][2])
	}
}

func TestParseCategoryRows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		rows          [][]any
		expectedLen   int
		checkPayee    string
		checkCategory *string
		checkName     *string
	}{
		{
			name:        "empty rows",
			rows:        [][]any{},
			expectedLen: 0,
		},
		{
			name: "single valid row",
			rows: [][]any{
				{"Starbucks", "Food", "Coffee"},
			},
			expectedLen:   1,
			checkPayee:    "Starbucks",
			checkCategory: strPtr("Food"),
			checkName:     strPtr("Coffee"),
		},
		{
			name: "multiple rows with some invalid",
			rows: [][]any{
				{"Store1", "Food"},
				{},
				{"Store2", "Transport", "Bus"},
				{123, "Invalid"},
			},
			expectedLen: 2,
		},
		{
			name: "row with empty payee",
			rows: [][]any{
				{"", "Food", "Name"},
			},
			expectedLen: 0,
		},
		{
			name: "row with non-string payee",
			rows: [][]any{
				{123, "Food", "Name"},
			},
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ParseCategoryRows(tt.rows)

			if len(result) != tt.expectedLen {
				t.Errorf("Expected %d entries, got %d", tt.expectedLen, len(result))
			}

			if tt.checkPayee != "" {
				cat, ok := result[tt.checkPayee]
				if !ok {
					t.Errorf("Expected payee '%s' in result", tt.checkPayee)
					return
				}

				if tt.checkCategory != nil {
					if cat.Category == nil || *cat.Category != *tt.checkCategory {
						t.Errorf("Expected category '%s', got %v", *tt.checkCategory, cat.Category)
					}
				}

				if tt.checkName != nil {
					if cat.Name == nil || *cat.Name != *tt.checkName {
						t.Errorf("Expected name '%s', got %v", *tt.checkName, cat.Name)
					}
				}
			}
		})
	}
}

func TestParseTransactionIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		rows        [][]any
		expectedLen int
		checkIDs    []string
	}{
		{
			name:        "empty rows",
			rows:        [][]any{},
			expectedLen: 0,
		},
		{
			name: "single ID",
			rows: [][]any{
				{"txn123"},
			},
			expectedLen: 1,
			checkIDs:    []string{"txn123"},
		},
		{
			name: "multiple IDs",
			rows: [][]any{
				{"txn1"},
				{"txn2"},
				{"txn3"},
			},
			expectedLen: 3,
			checkIDs:    []string{"txn1", "txn2", "txn3"},
		},
		{
			name: "with empty rows",
			rows: [][]any{
				{"txn1"},
				{},
				{"txn2"},
			},
			expectedLen: 2,
			checkIDs:    []string{"txn1", "txn2"},
		},
		{
			name: "with non-string IDs",
			rows: [][]any{
				{"txn1"},
				{123},
				{"txn2"},
			},
			expectedLen: 2,
			checkIDs:    []string{"txn1", "txn2"},
		},
		{
			name: "with extra columns",
			rows: [][]any{
				{"txn1", "extra", "data"},
				{"txn2", "more"},
			},
			expectedLen: 2,
			checkIDs:    []string{"txn1", "txn2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ParseTransactionIDs(tt.rows)

			if len(result) != tt.expectedLen {
				t.Errorf("Expected %d IDs, got %d", tt.expectedLen, len(result))
			}

			for _, id := range tt.checkIDs {
				if !result[id] {
					t.Errorf("Expected ID '%s' in result", id)
				}
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
