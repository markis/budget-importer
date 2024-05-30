from dataclasses import dataclass
from datetime import UTC, datetime
from decimal import Decimal
from typing import Any, Self, TypedDict, TypeGuard

from budget.models.paperless import Document


class SimpleFinOrganizationDict(TypedDict):
    domain: str
    name: str
    sfin_url: str


@dataclass
class SimpleFinOrganization:
    domain: str
    name: str
    sfin_url: str | None

    @classmethod
    def from_dict(cls, org: SimpleFinOrganizationDict) -> Self:
        return cls(
            domain=org["domain"],
            name=org["name"],
            sfin_url=org.get("sfin_url"),
        )


class SimpleFinHoldingDict(TypedDict):
    cost_basis: str
    created: int
    currency: str
    description: str
    id: str
    market_value: str
    purchase_price: str
    shares: str
    symbol: str


@dataclass
class SimpleFinHolding:
    cost_basis: str
    created: int
    currency: str
    description: str
    id: str
    market_value: str
    purchase_price: str
    shares: str
    symbol: str

    @classmethod
    def from_dict(cls, holding: SimpleFinHoldingDict) -> Self:
        return cls(
            cost_basis=holding["cost_basis"],
            created=holding["created"],
            currency=holding["currency"],
            description=holding["description"],
            id=holding["id"],
            market_value=holding["market_value"],
            purchase_price=holding["purchase_price"],
            shares=holding["shares"],
            symbol=holding["symbol"],
        )


class SimpleFinTransactionDict(TypedDict):
    id: str
    amount: str
    description: str
    memo: str
    payee: str
    posted: int
    transacted_at: int


@dataclass
class SimpleFinTransaction:
    id: str
    amount: Decimal
    description: str
    memo: str
    payee: str
    posted: datetime
    transacted_at: datetime
    category: str | None = None
    receipt: Document | None = None

    @classmethod
    def from_dict(cls, transaction: SimpleFinTransactionDict) -> Self:
        posted = datetime.fromtimestamp(transaction["posted"], tz=UTC)
        transacted_at = datetime.fromtimestamp(transaction["transacted_at"], tz=UTC)

        return cls(
            id=transaction["id"],
            amount=Decimal(transaction["amount"]),
            description=transaction["description"],
            memo=transaction["memo"],
            payee=transaction["payee"],
            posted=posted,
            transacted_at=transacted_at,
        )


class SimpleFinAccountDict(TypedDict("SimpleFinAccount", {"available-balance": str, "balance-date": int})):
    balance: str
    currency: str
    holdings: list[SimpleFinHoldingDict]
    id: str
    name: str
    org: SimpleFinOrganizationDict
    transactions: list[SimpleFinTransactionDict]


@dataclass
class SimpleFinAccount:
    available_balance: str
    balance: str
    balance_date: int
    currency: str
    holdings: list[SimpleFinHolding]
    id: str
    name: str
    org: SimpleFinOrganization
    transactions: list[SimpleFinTransaction]

    @classmethod
    def from_dict(cls, account: SimpleFinAccountDict) -> Self:
        org = SimpleFinOrganization.from_dict(account["org"])
        holdings = [SimpleFinHolding.from_dict(holding) for holding in account["holdings"]]
        transactions = [SimpleFinTransaction.from_dict(transaction) for transaction in account["transactions"]]
        return cls(
            available_balance=account["available-balance"],
            balance=account["balance"],
            balance_date=account["balance-date"],
            currency=account["currency"],
            holdings=holdings,
            id=account["id"],
            name=account["name"],
            org=org,
            transactions=transactions,
        )


class SimpleFinResponseDict(TypedDict):
    accounts: list[SimpleFinAccountDict]
    errors: list[str] | None
    x_api_message: list[str] | None


def is_simplefin_response(value: dict[str, Any] | Any) -> TypeGuard[SimpleFinResponseDict]:
    return isinstance(value, dict) and "accounts" in value


@dataclass
class SimpleFinResponse:
    accounts: list[SimpleFinAccount]
    errors: list[str] | None
    x_api_message: list[str] | None

    @classmethod
    def from_dict(cls, data: SimpleFinResponseDict) -> Self:
        accounts = [SimpleFinAccount.from_dict(account) for account in data["accounts"]]
        return cls(
            accounts=accounts,
            errors=data.get("errors"),
            x_api_message=data.get("x_api_message"),
        )
