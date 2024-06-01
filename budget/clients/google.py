import logging
from collections.abc import Sequence
from types import TracebackType
from typing import Self, TypeGuard

from gspread.auth import service_account
from gspread.client import Client
from gspread.utils import InsertDataOption, ValueInputOption

from budget.models.google import Category, GoogleSheetRow
from budget.models.simplefin import SimpleFinTransaction

logger = logging.getLogger(__name__)


def is_list_of_strings(data: list[list[str]]) -> TypeGuard[list[list[str]]]:
    return bool(data)


def convert_to_row(tran: SimpleFinTransaction) -> GoogleSheetRow:
    """Converts a SimpleFinTransaction to a row for Google Sheets."""
    return [
        tran.id,
        tran.payee,
        float(tran.amount),
        tran.transacted_at.strftime("%-m/%-d/%Y"),
        tran.category or "",
        str(tran.receipt) if tran.receipt else "",
    ]


class GoogleClient:
    google_client: Client

    def __init__(self, credentials: str) -> None:
        self.google_client = service_account(credentials)

    def __enter__(self) -> Self:
        return self

    def __exit__(
        self,
        exc_type: type[BaseException] | None,
        exc_val: BaseException | None,
        exc_tb: TracebackType | None,
    ) -> None:
        del exc_type, exc_val, exc_tb
        self.google_client.http_client.session.close()

    def get_category_mapping(self, spreadsheet_id: str, sheet_name: str) -> tuple[set[str], dict[str, Category]]:
        """Returns a mapping of transaction descriptions to categories."""
        sheet = self.google_client.open_by_key(spreadsheet_id)
        ws = sheet.worksheet(sheet_name)
        values = ws.get_all_values()
        assert is_list_of_strings(values)
        categories = {row[0] for row in values}
        mapping = {row[0]: Category.from_row(row) for row in values}
        return categories, mapping

    def insert_records_to_google_sheet(
        self, spreadsheet_id: str, sheet_name: str, transactions: Sequence[SimpleFinTransaction]
    ) -> None:
        """Inserts records into the Google Sheet."""
        sheet = self.google_client.open_by_key(spreadsheet_id)
        ws = sheet.worksheet(sheet_name)
        values = ws.get_all_values()
        assert is_list_of_strings(values)
        current_ids = {row[0] for row in values}
        records = [convert_to_row(transaction) for transaction in transactions if transaction.id not in current_ids]
        logger.info("Inserting %d records into Google Sheet", len(records))

        _ = ws.append_rows(
            records,
            insert_data_option=InsertDataOption.insert_rows,
            value_input_option=ValueInputOption.user_entered,
            include_values_in_response=True,
        )
        _ = ws.sort((4, "des"))
