// Package clients provides HTTP clients for external services.
package clients

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"github.com/markis/budget-importer/internal/models"
)

// GoogleSheetsClient handles communication with Google Sheets API.
type GoogleSheetsClient struct {
	service       *sheets.Service
	spreadsheetID string
}

// NewGoogleSheetsClient creates a new Google Sheets client.
func NewGoogleSheetsClient(
	ctx context.Context,
	credentialsFile, spreadsheetID string,
) (*GoogleSheetsClient, error) {
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	return &GoogleSheetsClient{
		service:       srv,
		spreadsheetID: spreadsheetID,
	}, nil
}

// GetCategoryMapping reads the category mapping from the lookup sheet.
func (c *GoogleSheetsClient) GetCategoryMapping(sheetName string) (map[string]models.Category, error) {
	readRange := sheetName + "!A:C"

	resp, err := c.service.Spreadsheets.Values.Get(c.spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read category mapping: %w", err)
	}

	mapping := make(map[string]models.Category)

	for _, row := range resp.Values {
		cat := parseRow(row)
		if cat == nil {
			continue
		}

		payee, ok := row[0].(string)
		if ok && payee != "" {
			mapping[payee] = *cat
		}
	}

	return mapping, nil
}

func parseRow(row []any) *models.Category {
	if len(row) == 0 {
		return nil
	}

	var category, name *string

	if len(row) > 1 {
		if cat, ok := row[1].(string); ok && cat != "" {
			category = &cat
		}
	}

	if len(row) > 2 {
		if n, ok := row[2].(string); ok && n != "" {
			name = &n
		}
	}

	return &models.Category{
		Category: category,
		Name:     name,
	}
}

// ParseCategoryRows converts raw sheet rows into a category mapping.
func ParseCategoryRows(rows [][]any) map[string]models.Category {
	mapping := make(map[string]models.Category)

	for _, row := range rows {
		cat := parseRow(row)
		if cat == nil {
			continue
		}

		payee, ok := row[0].(string)
		if ok && payee != "" {
			mapping[payee] = *cat
		}
	}

	return mapping
}

// ParseTransactionIDs extracts transaction IDs from sheet rows.
func ParseTransactionIDs(rows [][]any) map[string]bool {
	ids := make(map[string]bool)

	for _, row := range rows {
		if len(row) > 0 {
			if id, ok := row[0].(string); ok {
				ids[id] = true
			}
		}
	}

	return ids
}

// GetExistingTransactionIDs reads all existing transaction IDs from the sheet.
func (c *GoogleSheetsClient) GetExistingTransactionIDs(sheetName string) (map[string]bool, error) {
	readRange := sheetName + "!A:A"

	resp, err := c.service.Spreadsheets.Values.Get(c.spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read existing transaction IDs: %w", err)
	}

	ids := make(map[string]bool)

	for _, row := range resp.Values {
		if len(row) > 0 {
			if id, ok := row[0].(string); ok {
				ids[id] = true
			}
		}
	}

	return ids, nil
}

// AppendRows appends rows to the specified sheet.
func (c *GoogleSheetsClient) AppendRows(sheetName string, rows []models.GoogleSheetRow) error {
	if len(rows) == 0 {
		return nil
	}

	writeRange := sheetName + "!A:F"

	values := make([][]any, len(rows))
	for i, row := range rows {
		values[i] = row
	}

	valueRange := &sheets.ValueRange{
		Values: values,
	}

	_, err := c.service.Spreadsheets.Values.Append(c.spreadsheetID, writeRange, valueRange).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Do()
	if err != nil {
		return fmt.Errorf("failed to append rows: %w", err)
	}

	return nil
}

// SortByDate sorts the sheet by the date column (column D) in descending order.
func (c *GoogleSheetsClient) SortByDate(sheetName string) error {
	spreadsheet, err := c.service.Spreadsheets.Get(c.spreadsheetID).Do()
	if err != nil {
		return fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	var sheetID int64

	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			sheetID = sheet.Properties.SheetId

			break
		}
	}

	sortRequest := &sheets.Request{
		SortRange: &sheets.SortRangeRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetID,
				StartRowIndex:    1, // Skip header row.
				StartColumnIndex: 0,
				EndColumnIndex:   6,
			},
			SortSpecs: []*sheets.SortSpec{
				{
					DimensionIndex: 3, // Column D (date).
					SortOrder:      "DESCENDING",
				},
			},
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{sortRequest},
	}

	_, err = c.service.Spreadsheets.BatchUpdate(c.spreadsheetID, batchUpdateRequest).Do()
	if err != nil {
		return fmt.Errorf("failed to sort sheet: %w", err)
	}

	return nil
}
