from dataclasses import dataclass
from datetime import date
from decimal import Decimal
from enum import IntEnum
from typing import Any, NotRequired, Self, TypedDict, TypeGuard, override


class CustomFieldDict(TypedDict):
    value: str
    field: int


class DocumentDict(TypedDict):
    id: int
    document_type: int
    title: str
    content: str
    tags: list[str]
    created: str
    created_date: str
    custom_fields: list[CustomFieldDict]


class ResponseDict(TypedDict):
    count: int
    next: str | None
    all: NotRequired[list[int]]
    results: list[DocumentDict]


def is_response_dict(data: Any | ResponseDict) -> TypeGuard[ResponseDict]:
    return isinstance(data, dict) and all(key in data for key in ResponseDict.__required_keys__)


class CustomField(IntEnum):
    TOTAL = 1
    CATEGORY = 3


@dataclass
class Document:
    id: int
    date: date
    total: Decimal | None
    title: str
    category: str | None

    @override
    def __str__(self) -> str:
        return f"https://paperless.markis.network/documents/{self.id}/"

    @classmethod
    def from_dict(cls, data: DocumentDict) -> Self:
        """
        Create a Document instance from a dictionary.

        document_type: 1 = Receipt

        .. note::
        {
            "id": 18,
            "correspondent": null,
            "document_type": 1,
            "storage_path": null,
            "title": "Company Name",
            "content": "",
            "tags": [],
            "created": "2024-01-01T00:00:00-04:00",
            "created_date": "2024-01-01",
            "modified": "2024-01-2T12:30:21.123456-04:00",
            "added": "2024-01-3T12:30:10.123456-04:00",
            "archive_serial_number": null,
            "original_file_name": "document.pdf",
            "archived_file_name": "2024-01-01 Company Name.pdf",
            "owner": 3,
            "user_can_change": true,
            "is_shared_by_requester": false,
            "notes": [],
            "custom_fields": [
                {
                    "value": "USD16.75",
                    "field": 1
                },
                {
                    "value": "Category",
                    "field": 3
                }
            ]
        }
        """
        created = date.fromisoformat(data["created_date"])
        total: Decimal | None = None
        total_field = next((field for field in data["custom_fields"] if field["field"] == CustomField.TOTAL), None)
        if total_field is not None:
            total = -Decimal(total_field["value"].replace("USD", ""))
        category = next(
            (field["value"] for field in data["custom_fields"] if field["field"] == CustomField.CATEGORY), None
        )
        return cls(
            id=int(data["id"]),
            date=created,
            total=total,
            title=data["title"],
            category=category,
        )
