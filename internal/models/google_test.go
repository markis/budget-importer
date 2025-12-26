package models_test

import (
	"testing"

	"github.com/markis/budget-importer/internal/models"
)

func TestCategory(t *testing.T) {
	t.Parallel()

	t.Run("with all fields", func(t *testing.T) {
		t.Parallel()

		category := "Food"
		name := "Grocery Store"

		cat := models.Category{
			Category: &category,
			Name:     &name,
		}

		if cat.Category == nil || *cat.Category != "Food" {
			t.Errorf("Expected category 'Food', got %v", cat.Category)
		}

		if cat.Name == nil || *cat.Name != "Grocery Store" {
			t.Errorf("Expected name 'Grocery Store', got %v", cat.Name)
		}
	})

	t.Run("with nil fields", func(t *testing.T) {
		t.Parallel()

		cat := models.Category{
			Category: nil,
			Name:     nil,
		}

		if cat.Category != nil {
			t.Errorf("Expected nil category, got %v", cat.Category)
		}

		if cat.Name != nil {
			t.Errorf("Expected nil name, got %v", cat.Name)
		}
	})

	t.Run("with only category", func(t *testing.T) {
		t.Parallel()

		category := "Entertainment"

		cat := models.Category{
			Category: &category,
			Name:     nil,
		}

		if cat.Category == nil || *cat.Category != "Entertainment" {
			t.Errorf("Expected category 'Entertainment', got %v", cat.Category)
		}

		if cat.Name != nil {
			t.Errorf("Expected nil name, got %v", cat.Name)
		}
	})
}

func TestGoogleSheetRow(t *testing.T) {
	t.Parallel()

	t.Run("create row with mixed types", func(t *testing.T) {
		t.Parallel()

		row := models.GoogleSheetRow{
			"txn123",
			"Coffee Shop",
			-4.50,
			"12/26/2023",
			"Food",
			"",
		}

		if len(row) != 6 {
			t.Errorf("Expected 6 elements, got %d", len(row))
		}

		if row[0] != "txn123" {
			t.Errorf("Expected row[0] = 'txn123', got %v", row[0])
		}

		if row[2] != -4.50 {
			t.Errorf("Expected row[2] = -4.50, got %v", row[2])
		}
	})

	t.Run("empty row", func(t *testing.T) {
		t.Parallel()

		row := models.GoogleSheetRow{}

		if len(row) != 0 {
			t.Errorf("Expected 0 elements, got %d", len(row))
		}
	})

	t.Run("row with nil values", func(t *testing.T) {
		t.Parallel()

		row := models.GoogleSheetRow{
			"txn123",
			nil,
			0.0,
			"",
			nil,
			nil,
		}

		if len(row) != 6 {
			t.Errorf("Expected 6 elements, got %d", len(row))
		}

		if row[1] != nil {
			t.Errorf("Expected row[1] = nil, got %v", row[1])
		}
	})
}
