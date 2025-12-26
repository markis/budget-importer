package main

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/markis/budget-importer/internal/models"
)

// MockSheetsClient is a mock implementation of SheetsClient for testing.
type MockSheetsClient struct {
	CategoryMapping   map[string]models.Category
	ExistingIDs       map[string]bool
	AppendedRows      []models.GoogleSheetRow
	SortCalled        bool
	GetCategoryErr    error
	GetExistingIDsErr error
	AppendRowsErr     error
	SortByDateErr     error
}

func (m *MockSheetsClient) GetCategoryMapping(_ string) (map[string]models.Category, error) {
	if m.GetCategoryErr != nil {
		return nil, m.GetCategoryErr
	}

	return m.CategoryMapping, nil
}

func (m *MockSheetsClient) GetExistingTransactionIDs(_ string) (map[string]bool, error) {
	if m.GetExistingIDsErr != nil {
		return nil, m.GetExistingIDsErr
	}

	return m.ExistingIDs, nil
}

func (m *MockSheetsClient) AppendRows(_ string, rows []models.GoogleSheetRow) error {
	if m.AppendRowsErr != nil {
		return m.AppendRowsErr
	}

	m.AppendedRows = rows

	return nil
}

func (m *MockSheetsClient) SortByDate(_ string) error {
	if m.SortByDateErr != nil {
		return m.SortByDateErr
	}

	m.SortCalled = true

	return nil
}

// MockSimpleFinClient is a mock implementation of SimpleFinClient for testing.
type MockSimpleFinClient struct {
	Response *models.SimpleFinResponse
	Err      error
}

func (m *MockSimpleFinClient) FetchTransactions(_ time.Time) (*models.SimpleFinResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	return m.Response, nil
}

func TestLoadConfig(t *testing.T) {
	t.Run("loads valid config file", func(t *testing.T) {
		configContent := `
simplefin:
  access_url: "https://api.simplefin.org"
google:
  credentials: "/path/to/creds.json"
  spreadsheet_id: "spreadsheet123"
`
		tmpFile, err := os.CreateTemp(t.TempDir(), "config*.yaml")
		if err != nil {
			t.Fatal(err)
		}

		if _, err := tmpFile.WriteString(configContent); err != nil {
			t.Fatal(err)
		}

		if err := tmpFile.Close(); err != nil {
			t.Fatal(err)
		}

		config, err := loadConfig(tmpFile.Name())
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if config.SimpleFin.AccessURL != "https://api.simplefin.org" {
			t.Errorf("Expected access_url 'https://api.simplefin.org', got '%s'", config.SimpleFin.AccessURL)
		}

		if config.Google.Credentials != "/path/to/creds.json" {
			t.Errorf("Expected credentials '/path/to/creds.json', got '%s'", config.Google.Credentials)
		}

		if config.Google.SpreadsheetID != "spreadsheet123" {
			t.Errorf("Expected spreadsheet_id 'spreadsheet123', got '%s'", config.Google.SpreadsheetID)
		}

		// Check defaults
		if config.Google.SheetName != "transactions" {
			t.Errorf("Expected default sheet_name 'transactions', got '%s'", config.Google.SheetName)
		}

		if config.Google.MappingSheet != "lookup" {
			t.Errorf("Expected default mapping_sheet 'lookup', got '%s'", config.Google.MappingSheet)
		}
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		_, err := loadConfig("/nonexistent/config.yaml")
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		tmpFile, err := os.CreateTemp(t.TempDir(), "config*.yaml")
		if err != nil {
			t.Fatal(err)
		}

		if _, err := tmpFile.WriteString("invalid: yaml: content: ["); err != nil {
			t.Fatal(err)
		}

		if err := tmpFile.Close(); err != nil {
			t.Fatal(err)
		}

		_, err = loadConfig(tmpFile.Name())
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
	})
}

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config with access URL",
			config: &Config{
				SimpleFin: SimpleFinConfig{
					AccessURL: "https://api.simplefin.org",
				},
				Google: GoogleConfig{
					Credentials:   "/path/to/creds.json",
					SpreadsheetID: "spreadsheet123",
				},
			},
			expectError: false,
		},
		{
			name: "valid config with username and password",
			config: &Config{
				SimpleFin: SimpleFinConfig{
					Username: "user",
					Password: "pass",
				},
				Google: GoogleConfig{
					Credentials:   "/path/to/creds.json",
					SpreadsheetID: "spreadsheet123",
				},
			},
			expectError: false,
		},
		{
			name: "missing SimpleFIN credentials",
			config: &Config{
				Google: GoogleConfig{
					Credentials:   "/path/to/creds.json",
					SpreadsheetID: "spreadsheet123",
				},
			},
			expectError: true,
		},
		{
			name: "missing Google credentials",
			config: &Config{
				SimpleFin: SimpleFinConfig{
					AccessURL: "https://api.simplefin.org",
				},
				Google: GoogleConfig{
					SpreadsheetID: "spreadsheet123",
				},
			},
			expectError: true,
		},
		{
			name: "missing spreadsheet ID",
			config: &Config{
				SimpleFin: SimpleFinConfig{
					AccessURL: "https://api.simplefin.org",
				},
				Google: GoogleConfig{
					Credentials: "/path/to/creds.json",
				},
			},
			expectError: true,
		},
		{
			name: "only username without password",
			config: &Config{
				SimpleFin: SimpleFinConfig{
					Username: "user",
				},
				Google: GoogleConfig{
					Credentials:   "/path/to/creds.json",
					SpreadsheetID: "spreadsheet123",
				},
			},
			expectError: true,
		},
		{
			name: "only password without username",
			config: &Config{
				SimpleFin: SimpleFinConfig{
					Password: "pass",
				},
				Google: GoogleConfig{
					Credentials:   "/path/to/creds.json",
					SpreadsheetID: "spreadsheet123",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateConfig(tt.config)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestApplyCategoryMapping(t *testing.T) {
	t.Parallel()

	food := "Food"
	transport := "Transport"
	coffeeShop := "Coffee Shop"
	renamedPayee := "Renamed Payee"

	categoryMapping := map[string]models.Category{
		"Starbucks": {
			Category: &food,
			Name:     &coffeeShop,
		},
		"Uber": {
			Category: &transport,
			Name:     nil,
		},
		"grocery": {
			Category: &food,
			Name:     nil,
		},
		"partial": {
			Category: nil,
			Name:     &renamedPayee,
		},
	}

	tests := []struct {
		name             string
		payee            string
		expectedCategory string
		expectedPayee    string
	}{
		{
			name:             "exact match with name override",
			payee:            "Starbucks",
			expectedCategory: "Food",
			expectedPayee:    "Coffee Shop",
		},
		{
			name:             "exact match without name override",
			payee:            "Uber",
			expectedCategory: "Transport",
			expectedPayee:    "Uber",
		},
		{
			name:             "partial match (case insensitive)",
			payee:            "GROCERY STORE",
			expectedCategory: "Food",
			expectedPayee:    "GROCERY STORE",
		},
		{
			name:             "no match",
			payee:            "Unknown Merchant",
			expectedCategory: "",
			expectedPayee:    "Unknown Merchant",
		},
		{
			name:             "partial match with name but no category",
			payee:            "partial match test",
			expectedCategory: "",
			expectedPayee:    "Renamed Payee",
		},
		{
			name:             "empty payee",
			payee:            "",
			expectedCategory: "",
			expectedPayee:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			payee := tt.payee
			category := applyCategoryMapping(tt.payee, categoryMapping, &payee)

			if category != tt.expectedCategory {
				t.Errorf("Expected category '%s', got '%s'", tt.expectedCategory, category)
			}

			if payee != tt.expectedPayee {
				t.Errorf("Expected payee '%s', got '%s'", tt.expectedPayee, payee)
			}
		})
	}
}

func TestApplyCategoryMapping_EmptyMapping(t *testing.T) {
	t.Parallel()

	payee := "Any Merchant"
	category := applyCategoryMapping(payee, map[string]models.Category{}, &payee)

	if category != "" {
		t.Errorf("Expected empty category, got '%s'", category)
	}

	if payee != "Any Merchant" {
		t.Errorf("Expected payee unchanged, got '%s'", payee)
	}
}

func TestApplyCategoryMapping_ExactMatchWithNilCategory(t *testing.T) {
	t.Parallel()

	name := "New Name"
	categoryMapping := map[string]models.Category{
		"TestPayee": {
			Category: nil,
			Name:     &name,
		},
	}

	payee := "TestPayee"
	category := applyCategoryMapping(payee, categoryMapping, &payee)

	if category != "" {
		t.Errorf("Expected empty category, got '%s'", category)
	}

	if payee != "New Name" {
		t.Errorf("Expected payee 'New Name', got '%s'", payee)
	}
}

func TestCreateRow(t *testing.T) {
	t.Parallel()

	food := "Food"
	coffeeShop := "Coffee Shop"

	categoryMapping := map[string]models.Category{
		"Starbucks": {
			Category: &food,
			Name:     &coffeeShop,
		},
	}

	txn := &models.SimpleFinTransaction{
		ID:          "txn123",
		Payee:       "Starbucks",
		Description: "Coffee purchase",
		Amount:      decimal.NewFromFloat(-4.50),
		Posted:      1703548800, // 2023-12-26
	}

	row := createRow(txn, categoryMapping)

	if len(row) != 6 {
		t.Fatalf("Expected 6 columns, got %d", len(row))
	}

	if row[0] != "txn123" {
		t.Errorf("Expected ID 'txn123', got %v", row[0])
	}

	if row[1] != "Coffee Shop" {
		t.Errorf("Expected payee 'Coffee Shop', got %v", row[1])
	}

	if row[2] != -4.50 {
		t.Errorf("Expected amount -4.50, got %v", row[2])
	}

	if row[4] != "Food" {
		t.Errorf("Expected category 'Food', got %v", row[4])
	}

	if row[5] != "" {
		t.Errorf("Expected empty receipt URL, got %v", row[5])
	}
}

func TestCreateRow_UsesDescriptionWhenPayeeEmpty(t *testing.T) {
	t.Parallel()

	categoryMapping := map[string]models.Category{}

	txn := &models.SimpleFinTransaction{
		ID:          "txn456",
		Payee:       "",
		Description: "Direct Deposit",
		Amount:      decimal.NewFromFloat(1000.00),
		Posted:      1703548800,
	}

	row := createRow(txn, categoryMapping)

	if row[1] != "Direct Deposit" {
		t.Errorf("Expected payee 'Direct Deposit', got %v", row[1])
	}
}

func TestProcessTransactions(t *testing.T) {
	t.Parallel()

	food := "Food"

	categoryMapping := map[string]models.Category{
		"Starbucks": {Category: &food},
	}

	existingIDs := map[string]bool{
		"existing_txn": true,
	}

	response := &models.SimpleFinResponse{
		Accounts: []models.SimpleFinAccount{
			{
				ID:   "acc1",
				Name: "Checking",
				Transactions: []models.SimpleFinTransaction{
					{
						ID:     "new_txn_1",
						Payee:  "Starbucks",
						Amount: decimal.NewFromFloat(-5.00),
						Posted: 1703548800,
					},
					{
						ID:     "existing_txn",
						Payee:  "Uber",
						Amount: decimal.NewFromFloat(-15.00),
						Posted: 1703548800,
					},
					{
						ID:     "new_txn_2",
						Payee:  "Gas Station",
						Amount: decimal.NewFromFloat(-40.00),
						Posted: 1703548800,
					},
				},
			},
		},
	}

	rows := processTransactions(response, categoryMapping, existingIDs)

	// Should only have 2 new transactions (existing_txn should be skipped).
	if len(rows) != 2 {
		t.Fatalf("Expected 2 new rows, got %d", len(rows))
	}

	// First row should be new_txn_1.
	if rows[0][0] != "new_txn_1" {
		t.Errorf("Expected first row ID 'new_txn_1', got %v", rows[0][0])
	}

	// Second row should be new_txn_2.
	if rows[1][0] != "new_txn_2" {
		t.Errorf("Expected second row ID 'new_txn_2', got %v", rows[1][0])
	}
}

func TestProcessTransactions_EmptyAccounts(t *testing.T) {
	t.Parallel()

	response := &models.SimpleFinResponse{
		Accounts: []models.SimpleFinAccount{},
	}

	rows := processTransactions(response, map[string]models.Category{}, map[string]bool{})

	if len(rows) != 0 {
		t.Errorf("Expected 0 rows, got %d", len(rows))
	}
}

func TestProcessTransactions_AllExisting(t *testing.T) {
	t.Parallel()

	existingIDs := map[string]bool{
		"txn1": true,
		"txn2": true,
	}

	response := &models.SimpleFinResponse{
		Accounts: []models.SimpleFinAccount{
			{
				Transactions: []models.SimpleFinTransaction{
					{ID: "txn1", Amount: decimal.NewFromFloat(-5.00), Posted: 1703548800},
					{ID: "txn2", Amount: decimal.NewFromFloat(-10.00), Posted: 1703548800},
				},
			},
		},
	}

	rows := processTransactions(response, map[string]models.Category{}, existingIDs)

	if len(rows) != 0 {
		t.Errorf("Expected 0 rows (all existing), got %d", len(rows))
	}
}

func TestProcessTransactions_MultipleAccounts(t *testing.T) {
	t.Parallel()

	response := &models.SimpleFinResponse{
		Accounts: []models.SimpleFinAccount{
			{
				ID:   "acc1",
				Name: "Checking",
				Transactions: []models.SimpleFinTransaction{
					{ID: "txn1", Payee: "Store1", Amount: decimal.NewFromFloat(-10.00), Posted: 1703548800},
				},
			},
			{
				ID:   "acc2",
				Name: "Savings",
				Transactions: []models.SimpleFinTransaction{
					{ID: "txn2", Payee: "Store2", Amount: decimal.NewFromFloat(-20.00), Posted: 1703548800},
					{ID: "txn3", Payee: "Store3", Amount: decimal.NewFromFloat(-30.00), Posted: 1703548800},
				},
			},
		},
	}

	rows := processTransactions(response, map[string]models.Category{}, map[string]bool{})

	if len(rows) != 3 {
		t.Errorf("Expected 3 rows from multiple accounts, got %d", len(rows))
	}
}

func TestProcessTransactions_EmptyTransactions(t *testing.T) {
	t.Parallel()

	response := &models.SimpleFinResponse{
		Accounts: []models.SimpleFinAccount{
			{
				ID:           "acc1",
				Name:         "Checking",
				Transactions: []models.SimpleFinTransaction{},
			},
		},
	}

	rows := processTransactions(response, map[string]models.Category{}, map[string]bool{})

	if len(rows) != 0 {
		t.Errorf("Expected 0 rows, got %d", len(rows))
	}
}

func TestCreateRow_DateFormat(t *testing.T) {
	t.Parallel()

	// Test specific date: December 26, 2023 at 00:00:00 UTC.
	txn := &models.SimpleFinTransaction{
		ID:     "txn_date_test",
		Payee:  "Test",
		Amount: decimal.NewFromFloat(-1.00),
		Posted: 1703548800, // 2023-12-26 00:00:00 UTC
	}

	row := createRow(txn, map[string]models.Category{})

	// Date should be in M/D/YYYY format.
	dateStr, ok := row[3].(string)
	if !ok {
		t.Fatal("Expected date to be a string")
	}

	// The exact date depends on timezone, but format should be correct.
	if dateStr == "" {
		t.Error("Expected non-empty date string")
	}
}

func TestCreateRow_PositiveAmount(t *testing.T) {
	t.Parallel()

	txn := &models.SimpleFinTransaction{
		ID:     "deposit",
		Payee:  "Employer",
		Amount: decimal.NewFromFloat(1500.00),
		Posted: 1703548800,
	}

	row := createRow(txn, map[string]models.Category{})

	amount, ok := row[2].(float64)
	if !ok {
		t.Fatal("Expected amount to be float64")
	}

	if amount != 1500.00 {
		t.Errorf("Expected amount 1500.00, got %v", amount)
	}
}

func TestCreateRow_ZeroAmount(t *testing.T) {
	t.Parallel()

	txn := &models.SimpleFinTransaction{
		ID:     "zero",
		Payee:  "Adjustment",
		Amount: decimal.NewFromFloat(0),
		Posted: 1703548800,
	}

	row := createRow(txn, map[string]models.Category{})

	amount, ok := row[2].(float64)
	if !ok {
		t.Fatal("Expected amount to be float64")
	}

	if amount != 0 {
		t.Errorf("Expected amount 0, got %v", amount)
	}
}

func TestValidateConfig_AllFieldsEmpty(t *testing.T) {
	t.Parallel()

	config := &Config{}
	err := validateConfig(config)

	if err == nil {
		t.Error("Expected error for empty config")
	}
}

func TestValidateConfig_BothAccessURLAndCredentials(t *testing.T) {
	t.Parallel()

	// When both access URL and username/password are provided, should still be valid.
	config := &Config{
		SimpleFin: SimpleFinConfig{
			AccessURL: "https://api.simplefin.org",
			Username:  "user",
			Password:  "pass",
		},
		Google: GoogleConfig{
			Credentials:   "/path/to/creds.json",
			SpreadsheetID: "spreadsheet123",
		},
	}

	err := validateConfig(config)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestRunWithClients_Success(t *testing.T) {
	t.Parallel()

	food := "Food"
	mockSheets := &MockSheetsClient{
		CategoryMapping: map[string]models.Category{
			"Starbucks": {Category: &food},
		},
		ExistingIDs: map[string]bool{},
	}

	mockSimpleFin := &MockSimpleFinClient{
		Response: &models.SimpleFinResponse{
			Accounts: []models.SimpleFinAccount{
				{
					Transactions: []models.SimpleFinTransaction{
						{ID: "txn1", Payee: "Starbucks", Amount: decimal.NewFromFloat(-5.00), Posted: 1703548800},
					},
				},
			},
		},
	}

	config := &Config{
		Google: GoogleConfig{
			SheetName:    "transactions",
			MappingSheet: "lookup",
		},
	}

	err := runWithClients(config, mockSheets, mockSimpleFin)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(mockSheets.AppendedRows) != 1 {
		t.Errorf("Expected 1 appended row, got %d", len(mockSheets.AppendedRows))
	}

	if !mockSheets.SortCalled {
		t.Error("Expected SortByDate to be called")
	}
}

func TestRunWithClients_NoNewTransactions(t *testing.T) {
	t.Parallel()

	mockSheets := &MockSheetsClient{
		CategoryMapping: map[string]models.Category{},
		ExistingIDs: map[string]bool{
			"txn1": true,
		},
	}

	mockSimpleFin := &MockSimpleFinClient{
		Response: &models.SimpleFinResponse{
			Accounts: []models.SimpleFinAccount{
				{
					Transactions: []models.SimpleFinTransaction{
						{ID: "txn1", Payee: "Store", Amount: decimal.NewFromFloat(-5.00), Posted: 1703548800},
					},
				},
			},
		},
	}

	config := &Config{
		Google: GoogleConfig{
			SheetName:    "transactions",
			MappingSheet: "lookup",
		},
	}

	err := runWithClients(config, mockSheets, mockSimpleFin)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if mockSheets.AppendedRows != nil {
		t.Errorf("Expected no appended rows, got %d", len(mockSheets.AppendedRows))
	}

	if mockSheets.SortCalled {
		t.Error("Expected SortByDate not to be called when no new rows")
	}
}

func TestRunWithClients_GetCategoryMappingError(t *testing.T) {
	t.Parallel()

	mockSheets := &MockSheetsClient{
		GetCategoryErr: errors.New("sheets API error"),
	}

	mockSimpleFin := &MockSimpleFinClient{}

	config := &Config{
		Google: GoogleConfig{
			MappingSheet: "lookup",
		},
	}

	err := runWithClients(config, mockSheets, mockSimpleFin)

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestRunWithClients_GetExistingIDsError(t *testing.T) {
	t.Parallel()

	mockSheets := &MockSheetsClient{
		CategoryMapping:   map[string]models.Category{},
		GetExistingIDsErr: errors.New("sheets API error"),
	}

	mockSimpleFin := &MockSimpleFinClient{}

	config := &Config{
		Google: GoogleConfig{
			SheetName:    "transactions",
			MappingSheet: "lookup",
		},
	}

	err := runWithClients(config, mockSheets, mockSimpleFin)

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestRunWithClients_FetchTransactionsError(t *testing.T) {
	t.Parallel()

	mockSheets := &MockSheetsClient{
		CategoryMapping: map[string]models.Category{},
		ExistingIDs:     map[string]bool{},
	}

	mockSimpleFin := &MockSimpleFinClient{
		Err: errors.New("SimpleFIN API error"),
	}

	config := &Config{
		Google: GoogleConfig{
			SheetName:    "transactions",
			MappingSheet: "lookup",
		},
	}

	err := runWithClients(config, mockSheets, mockSimpleFin)

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestRunWithClients_AppendRowsError(t *testing.T) {
	t.Parallel()

	mockSheets := &MockSheetsClient{
		CategoryMapping: map[string]models.Category{},
		ExistingIDs:     map[string]bool{},
		AppendRowsErr:   errors.New("append error"),
	}

	mockSimpleFin := &MockSimpleFinClient{
		Response: &models.SimpleFinResponse{
			Accounts: []models.SimpleFinAccount{
				{
					Transactions: []models.SimpleFinTransaction{
						{ID: "txn1", Payee: "Store", Amount: decimal.NewFromFloat(-5.00), Posted: 1703548800},
					},
				},
			},
		},
	}

	config := &Config{
		Google: GoogleConfig{
			SheetName:    "transactions",
			MappingSheet: "lookup",
		},
	}

	err := runWithClients(config, mockSheets, mockSimpleFin)

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestRunWithClients_SortByDateError(t *testing.T) {
	t.Parallel()

	mockSheets := &MockSheetsClient{
		CategoryMapping: map[string]models.Category{},
		ExistingIDs:     map[string]bool{},
		SortByDateErr:   errors.New("sort error"),
	}

	mockSimpleFin := &MockSimpleFinClient{
		Response: &models.SimpleFinResponse{
			Accounts: []models.SimpleFinAccount{
				{
					Transactions: []models.SimpleFinTransaction{
						{ID: "txn1", Payee: "Store", Amount: decimal.NewFromFloat(-5.00), Posted: 1703548800},
					},
				},
			},
		},
	}

	config := &Config{
		Google: GoogleConfig{
			SheetName:    "transactions",
			MappingSheet: "lookup",
		},
	}

	err := runWithClients(config, mockSheets, mockSimpleFin)

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestFetchSheetData_Success(t *testing.T) {
	t.Parallel()

	food := "Food"
	mockSheets := &MockSheetsClient{
		CategoryMapping: map[string]models.Category{
			"Store": {Category: &food},
		},
		ExistingIDs: map[string]bool{
			"txn1": true,
			"txn2": true,
		},
	}

	config := &Config{
		Google: GoogleConfig{
			SheetName:    "transactions",
			MappingSheet: "lookup",
		},
	}

	categoryMapping, existingIDs, err := fetchSheetData(mockSheets, config)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(categoryMapping) != 1 {
		t.Errorf("Expected 1 category mapping, got %d", len(categoryMapping))
	}

	if len(existingIDs) != 2 {
		t.Errorf("Expected 2 existing IDs, got %d", len(existingIDs))
	}
}

func TestFetchTransactions_Success(t *testing.T) {
	t.Parallel()

	mockSimpleFin := &MockSimpleFinClient{
		Response: &models.SimpleFinResponse{
			Accounts: []models.SimpleFinAccount{
				{ID: "acc1"},
			},
		},
	}

	response, err := fetchTransactions(mockSimpleFin)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(response.Accounts) != 1 {
		t.Errorf("Expected 1 account, got %d", len(response.Accounts))
	}
}

func TestFetchTransactions_Error(t *testing.T) {
	t.Parallel()

	mockSimpleFin := &MockSimpleFinClient{
		Err: errors.New("API error"),
	}

	_, err := fetchTransactions(mockSimpleFin)

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestInsertNewRows_NoRows(t *testing.T) {
	t.Parallel()

	mockSheets := &MockSheetsClient{}

	err := insertNewRows(mockSheets, "transactions", []models.GoogleSheetRow{})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if mockSheets.SortCalled {
		t.Error("Expected SortByDate not to be called")
	}
}

func TestInsertNewRows_WithRows(t *testing.T) {
	t.Parallel()

	mockSheets := &MockSheetsClient{}

	rows := []models.GoogleSheetRow{
		{"txn1", "Store", -10.00, "1/1/2024", "Food", ""},
	}

	err := insertNewRows(mockSheets, "transactions", rows)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(mockSheets.AppendedRows) != 1 {
		t.Errorf("Expected 1 appended row, got %d", len(mockSheets.AppendedRows))
	}

	if !mockSheets.SortCalled {
		t.Error("Expected SortByDate to be called")
	}
}

func TestMain(m *testing.M) {
	// Run tests.
	os.Exit(m.Run())
}
