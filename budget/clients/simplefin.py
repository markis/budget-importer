import http.client
import json
import logging
from base64 import b64encode
from collections import defaultdict
from collections.abc import Sequence
from datetime import datetime
from functools import cached_property
from types import TracebackType
from typing import TYPE_CHECKING, Final, Self
from urllib.parse import ParseResult, urlencode, urlparse

from budget.models.google import Category
from budget.models.paperless import Document
from budget.models.simplefin import (
    SimpleFinAccount,
    SimpleFinResponse,
    SimpleFinResponseDict,
    SimpleFinTransaction,
    is_simplefin_response,
)

if TYPE_CHECKING:
    from decimal import Decimal


logger = logging.getLogger(__name__)


class SimpleFinClient:
    """
    SimpleFin class to interact with the SimpleFin API

    The SimpleFin API is a simple API that provides access to financial data.

    Sample usage:
    ```python
    with Client() as client:
        data = client.fetch_data(setup_token)
    ```
    """

    username: Final[str]
    password: Final[str]
    url: Final[ParseResult]
    conn: http.client.HTTPConnection | http.client.HTTPSConnection

    def __init__(self, url: str, username: str, password: str) -> None:
        self.username = username
        self.password = password
        self.url = urlparse(url)
        self.conn = http.client.HTTPSConnection(self.url.netloc, self.url.port)

    def __enter__(self) -> Self:
        return self

    def __exit__(
        self,
        exc_type: type[BaseException] | None,
        exc_val: BaseException | None,
        exc_tb: TracebackType | None,
    ) -> None:
        del exc_type, exc_val, exc_tb
        self.conn.close()

    @cached_property
    def auth_headers(self) -> dict[str, str]:
        credentials = f"{self.username}:{self.password}"
        encoded_credentials = b64encode(credentials.encode()).decode("ascii")
        return {"Authorization": f"Basic {encoded_credentials}"}

    def fetch_data(self, start_date: datetime) -> list[SimpleFinAccount]:
        """
        Fetches data from the SimpleFin API.
        """
        unix_start_date = int(start_date.timestamp())
        encoded_params = urlencode({"pending": 1, "start-date": unix_start_date})
        path = f"{self.url.path}/accounts?{encoded_params}"

        self.conn.request("GET", path, headers=self.auth_headers)
        with self.conn.getresponse() as response:
            if response.status != http.client.OK:
                msg = f"Failed to get data: {response.status}"
                raise ValueError(msg)

            data: SimpleFinResponseDict = json.loads(response.read().decode())

        if not is_simplefin_response(data):
            msg = f"Invalid response: {data}"
            raise ValueError(msg)

        resp = SimpleFinResponse.from_dict(data)
        logger.info("Fetched %d accounts", len(resp.accounts))
        return resp.accounts

    def categorize_transactions(
        self, transactions: Sequence[SimpleFinTransaction], mapping: dict[str, Category]
    ) -> None:
        """
        Categorize transactions based on the mapping.
        """
        for transaction in transactions:
            category, name = mapping.get(transaction.payee, (None, None))
            if not transaction.category and category:
                transaction.category = category
            if name:
                transaction.payee = name

    def attach_receipts(
        self, accounts: Sequence[SimpleFinAccount], receipts: Sequence[Document]
    ) -> list[SimpleFinTransaction]:
        """
        Attach receipts to transactions.
        """
        grouped_receipts: defaultdict[Decimal, list[Document]] = defaultdict(list)
        for receipt in receipts:
            if receipt.total:
                grouped_receipts[receipt.total].append(receipt)

        transactions: list[SimpleFinTransaction] = []
        for account in accounts:
            for transaction in account.transactions:
                documents = grouped_receipts.get(transaction.amount, [])
                document = next(iter(sorted(documents, key=lambda d: transaction.transacted_at.date() - d.date)), None)
                transaction.category = document.category if document else None
                transaction.receipt = document
                transactions.append(transaction)

        transactions.sort(key=lambda t: t.transacted_at, reverse=True)
        logger.info("Attached receipts to %d transactions", len(transactions))
        return transactions
