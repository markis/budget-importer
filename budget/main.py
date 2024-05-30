import logging
from dataclasses import dataclass
from datetime import UTC, datetime, timedelta
from functools import cached_property

from budget.clients.google import GoogleClient
from budget.clients.paperless import PaperlessClient
from budget.clients.simplefin import SimpleFinClient

logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(message)s")
logger = logging.getLogger(__name__)
logger.setLevel(logging.INFO)


@dataclass()
class Args:
    class Error(Exception): ...

    simplefin_username: str
    simplefin_password: str
    simplefin_access_url: str
    paperless_url: str
    paperless_token: str
    google_credentials: str
    sheets_spreadsheet_id: str
    sheets_range_name: str
    mapping_range_name: str

    @cached_property
    def start_date(self) -> datetime:
        return datetime.now(UTC) - timedelta(days=2)

    def __post_init__(self) -> None:
        errors: list[str] = []
        if not any((self.simplefin_username, self.simplefin_password, self.simplefin_access_url)):
            errors.append("SimpleFin credentials are required")
        if not any((self.paperless_url, self.paperless_token)):
            errors.append("Paperless credentials are required")
        if not any((self.google_credentials, self.sheets_spreadsheet_id)):
            errors.append("Google credentials are required")

        if errors:
            msg = f"Missing CLI Args \n{'\n'.join(errors)}"
            raise Args.Error(msg)


def main(args: Args) -> None:
    with (
        PaperlessClient(args.paperless_url, args.paperless_token) as paperless,
        SimpleFinClient(args.simplefin_access_url, args.simplefin_username, args.simplefin_password) as simplefin,
        GoogleClient(args.google_credentials) as google,
    ):
        _, mapping = google.get_category_mapping(args.sheets_spreadsheet_id, args.mapping_range_name)

        documents = paperless.fetch_documents()
        accounts = simplefin.fetch_data(args.start_date)

        transactions = simplefin.attach_receipts(accounts, documents)
        simplefin.categorize_transactions(transactions, mapping)

        google.insert_records_to_google_sheet(args.sheets_spreadsheet_id, args.sheets_range_name, transactions)
