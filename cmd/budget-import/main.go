// Package main provides the budget-import CLI tool that imports
// financial transactions from SimpleFIN into Google Sheets.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"

	"github.com/markis/budget-importer/internal/clients"
	"github.com/markis/budget-importer/internal/models"
)

// SheetsClient defines the interface for Google Sheets operations.
type SheetsClient interface {
	GetCategoryMapping(sheetName string) (map[string]models.Category, error)
	GetExistingTransactionIDs(sheetName string) (map[string]bool, error)
	AppendRows(sheetName string, rows []models.GoogleSheetRow) error
	SortByDate(sheetName string) error
}

// SimpleFinClient defines the interface for SimpleFIN operations.
type SimpleFinClient interface {
	FetchTransactions(startDate time.Time) (*models.SimpleFinResponse, error)
}

// Config holds the application configuration.
type Config struct {
	// SimpleFIN settings.
	SimpleFin SimpleFinConfig `yaml:"simplefin"`

	// Google Sheets settings.
	Google GoogleConfig `yaml:"google"`
}

// SimpleFinConfig holds SimpleFIN API configuration.
type SimpleFinConfig struct {
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	AccessURL string `yaml:"access_url"`
}

// GoogleConfig holds Google Sheets configuration.
type GoogleConfig struct {
	Credentials   string `yaml:"credentials"`
	SpreadsheetID string `yaml:"spreadsheet_id"`
	SheetName     string `default:"transactions" yaml:"sheet_name"`
	MappingSheet  string `default:"lookup"       yaml:"mapping_sheet"`
}

func main() {
	config := parseConfig()

	if err := run(config); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func parseConfig() *Config {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")

	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := validateConfig(config); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	return config
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	if err := defaults.Set(config); err != nil {
		return nil, fmt.Errorf("failed to set config defaults: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

func validateConfig(config *Config) error {
	hasSimpleFin := config.SimpleFin.AccessURL != "" ||
		(config.SimpleFin.Username != "" && config.SimpleFin.Password != "")
	if !hasSimpleFin {
		return fmt.Errorf("SimpleFIN credentials required: provide access_url or username/password")
	}

	if config.Google.Credentials == "" {
		return fmt.Errorf("google credentials path required")
	}

	if config.Google.SpreadsheetID == "" {
		return fmt.Errorf("google Sheets spreadsheet ID required")
	}

	return nil
}

func run(config *Config) error {
	ctx := context.Background()

	sheetsClient, err := initSheetsClient(ctx, config)
	if err != nil {
		return err
	}

	simpleFinClient := clients.NewSimpleFinClient(
		config.SimpleFin.AccessURL,
		config.SimpleFin.Username,
		config.SimpleFin.Password,
	)

	return runWithClients(config, sheetsClient, simpleFinClient)
}

func runWithClients(config *Config, sheetsClient SheetsClient, simpleFinClient SimpleFinClient) error {
	categoryMapping, existingIDs, err := fetchSheetData(sheetsClient, config)
	if err != nil {
		return err
	}

	response, err := fetchTransactions(simpleFinClient)
	if err != nil {
		return err
	}

	newRows := processTransactions(response, categoryMapping, existingIDs)

	return insertNewRows(sheetsClient, config.Google.SheetName, newRows)
}

func initSheetsClient(ctx context.Context, config *Config) (*clients.GoogleSheetsClient, error) {
	log.Println("Connecting to Google Sheets...")

	sheetsClient, err := clients.NewGoogleSheetsClient(
		ctx,
		config.Google.Credentials,
		config.Google.SpreadsheetID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Sheets client: %w", err)
	}

	return sheetsClient, nil
}

func fetchSheetData(
	sheetsClient SheetsClient,
	config *Config,
) (map[string]models.Category, map[string]bool, error) {
	log.Println("Fetching category mapping...")

	categoryMapping, err := sheetsClient.GetCategoryMapping(config.Google.MappingSheet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get category mapping: %w", err)
	}

	log.Printf("Loaded %d category mappings", len(categoryMapping))

	log.Println("Fetching existing transaction IDs...")

	existingIDs, err := sheetsClient.GetExistingTransactionIDs(config.Google.SheetName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get existing transaction IDs: %w", err)
	}

	log.Printf("Found %d existing transactions", len(existingIDs))

	return categoryMapping, existingIDs, nil
}

func fetchTransactions(simpleFinClient SimpleFinClient) (*models.SimpleFinResponse, error) {
	startDate := time.Now().AddDate(0, 0, -2)
	log.Printf("Fetching transactions since %s...", startDate.Format("2006-01-02"))

	response, err := simpleFinClient.FetchTransactions(startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	return response, nil
}

func processTransactions(
	response *models.SimpleFinResponse,
	categoryMapping map[string]models.Category,
	existingIDs map[string]bool,
) []models.GoogleSheetRow {
	var newRows []models.GoogleSheetRow

	var totalTransactions int

	for i := range response.Accounts {
		account := &response.Accounts[i]

		for j := range account.Transactions {
			txn := &account.Transactions[j]
			totalTransactions++

			if existingIDs[txn.ID] {
				continue
			}

			row := createRow(txn, categoryMapping)
			newRows = append(newRows, row)
		}
	}

	log.Printf("Found %d total transactions, %d new", totalTransactions, len(newRows))

	return newRows
}

func createRow(txn *models.SimpleFinTransaction, categoryMapping map[string]models.Category) models.GoogleSheetRow {
	payee := txn.Payee
	if payee == "" {
		payee = txn.Description
	}

	category := applyCategoryMapping(payee, categoryMapping, &payee)

	date := txn.PostedTime().Format("1/2/2006")
	amount, _ := txn.Amount.Float64()

	return models.GoogleSheetRow{
		txn.ID,
		payee,
		amount,
		date,
		category,
		"", // No receipt URL (paperless disabled).
	}
}

func applyCategoryMapping(payee string, categoryMapping map[string]models.Category, payeeOut *string) string {
	// Try exact match first.
	if cat, ok := categoryMapping[payee]; ok {
		if cat.Name != nil {
			*payeeOut = *cat.Name
		}

		if cat.Category != nil {
			return *cat.Category
		}
	}

	// Try partial match.
	for key, cat := range categoryMapping {
		if strings.Contains(strings.ToLower(payee), strings.ToLower(key)) {
			if cat.Name != nil {
				*payeeOut = *cat.Name
			}

			if cat.Category != nil {
				return *cat.Category
			}

			break
		}
	}

	return ""
}

func insertNewRows(
	sheetsClient SheetsClient,
	sheetName string,
	newRows []models.GoogleSheetRow,
) error {
	if len(newRows) == 0 {
		log.Println("No new transactions to insert")

		return nil
	}

	log.Printf("Inserting %d new transactions...", len(newRows))

	if err := sheetsClient.AppendRows(sheetName, newRows); err != nil {
		return fmt.Errorf("failed to append rows: %w", err)
	}

	log.Println("Sorting sheet by date...")

	if err := sheetsClient.SortByDate(sheetName); err != nil {
		return fmt.Errorf("failed to sort sheet: %w", err)
	}

	log.Println("Done!")

	return nil
}
