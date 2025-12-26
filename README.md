# Budget Importer

A CLI tool that imports financial transactions from SimpleFIN into Google Sheets for budgeting.

## Features

- Fetches transactions from SimpleFIN API
- Applies category mappings from a lookup sheet
- Inserts new transactions into Google Sheets (avoids duplicates)
- Sorts transactions by date

## Installation

```bash
go install github.com/markis/budget-importer/cmd/budget-import@latest
```

Or build from source:

```bash
go build -o budget-import ./cmd/budget-import
```

## Configuration

Configuration is provided via a YAML file. Copy `config.example.yaml` to `config.yaml` and fill in your values:

```yaml
simplefin:
  access_url: "https://beta-bridge.simplefin.org/simplefin"
  username: "your-username"
  password: "your-password"

google:
  credentials: "/path/to/service-account.json"
  spreadsheet_id: "your-spreadsheet-id"
  sheet_name: "transactions"      # default: "transactions"
  mapping_sheet: "lookup"         # default: "lookup"
```

### Configuration Options

| Field | Description |
|-------|-------------|
| `simplefin.access_url` | SimpleFIN access URL |
| `simplefin.username` | SimpleFIN username |
| `simplefin.password` | SimpleFIN password |
| `google.credentials` | Path to Google service account JSON |
| `google.spreadsheet_id` | Google Sheets spreadsheet ID |
| `google.sheet_name` | Sheet name for transactions (default: "transactions") |
| `google.mapping_sheet` | Sheet name for category mappings (default: "lookup") |

## Usage

```bash
# Using default config path (config.yaml)
budget-import

# Using custom config path
budget-import -config /path/to/config.yaml
```

## Google Sheets Format

### Transactions Sheet

| Column | Content |
|--------|---------|
| A | Transaction ID |
| B | Payee |
| C | Amount |
| D | Date (M/D/YYYY) |
| E | Category |
| F | Receipt URL |

### Lookup Sheet (Category Mapping)

| Column | Content |
|--------|---------|
| A | Payee name (key) |
| B | Category |
| C | Display name override (optional) |

## Docker

```bash
docker build -t budget-import .
docker run --rm \
  -v /path/to/config.yaml:/config.yaml:ro \
  -v /path/to/credentials.json:/credentials.json:ro \
  budget-import
```
