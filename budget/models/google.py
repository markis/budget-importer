from typing import NamedTuple, Self

GoogleSheetRow = list[str | float | int]


class Category(NamedTuple):
    category: str | None
    name: str | None

    @classmethod
    def from_row(cls, row: list[str]) -> Self:
        checked_row = [*row[1:], None, None]
        return cls(category=checked_row[0], name=checked_row[1])
