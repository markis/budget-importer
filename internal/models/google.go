// Package models provides data structures for the budget importer.
package models

// Category represents a category mapping from the lookup sheet.
type Category struct {
	Category *string // Category name.
	Name     *string // Display name override for payee.
}

// GoogleSheetRow represents a row to be inserted into Google Sheets.
type GoogleSheetRow []any
