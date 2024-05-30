import http.client
import json
import logging
from collections.abc import Generator
from functools import cached_property
from types import TracebackType
from typing import Final, Self
from urllib.parse import ParseResult, urlencode, urlparse

from budget.models.paperless import Document, ResponseDict, is_response_dict

logger = logging.getLogger(__name__)


class PaperlessClient:
    url: Final[ParseResult]
    token: Final[str]
    conn: http.client.HTTPConnection

    def __init__(self, url: str, token: str) -> None:
        self.token = token
        self.url = urlparse(url)
        hostname = self.url.hostname or self.url.netloc
        self.conn = http.client.HTTPConnection(hostname, self.url.port)

    def __enter__(self) -> Self:
        return self

    def __exit__(
        self,
        exc_type: type[BaseException] | None,
        exc_val: BaseException | None,
        exc_tb: TracebackType | None,
    ) -> None:
        del exc_type, exc_val, exc_tb  # unused
        self.conn.close()

    @cached_property
    def headers(self) -> dict[str, str]:
        return {
            "Accept": "application/json",
            "Accept-Encoding": "application/json",
            "Authorization": f"Token {self.token}",
        }

    def fetch_documents(self, document_type: str = "receipt") -> list[Document]:
        """Fetches documents from the Paperless API."""
        query = urlencode({"document_type__name__iexact": document_type})
        docs = list(self._inner_fetch_documents(f"/api/documents/?{query}"))
        logger.info("Fetched %d receipts", len(docs))
        return docs

    def _inner_fetch_documents(self, url: str) -> Generator[Document, None, None]:
        self.conn.request("GET", url, headers=self.headers)
        with self.conn.getresponse() as response:
            if response.status != http.client.OK:
                msg = f"Failed to get data: {response.status}"
                raise ValueError(msg)

            data: ResponseDict = json.loads(response.read().decode())

        if not is_response_dict(data):
            msg = f"Invalid response: {data}"
            raise ValueError(msg)

        for document_dict in data["results"]:
            yield Document.from_dict(document_dict)

        if data["next"]:
            yield from self._inner_fetch_documents(data["next"])

        return None
