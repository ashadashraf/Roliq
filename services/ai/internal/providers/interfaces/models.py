from __future__ import annotations

from dataclasses import dataclass
from enum import StrEnum
from typing import Any


class ProviderErrorCategory(StrEnum):
    VALIDATION = "validation"
    AUTHENTICATION = "authentication"
    RATE_LIMIT = "rate_limit"
    TIMEOUT = "timeout"
    UNAVAILABLE = "unavailable"
    BUDGET_EXCEEDED = "budget_exceeded"
    CONTENT_REJECTED = "content_rejected"
    PERMANENT_INPUT = "permanent_input"


class ProviderError(RuntimeError):
    def __init__(self, category: ProviderErrorCategory, code: str, retryable: bool):
        super().__init__(f"{category.value}:{code}")
        self.category = category
        self.code = code
        self.retryable = retryable


@dataclass(frozen=True)
class ProviderUsage:
    input_units: int = 0
    output_units: int = 0
    estimated_cost_usd: float = 0.0


@dataclass(frozen=True)
class ParseRequest:
    source_sha256: str
    schema_version: str
    prompt_version: str
    max_input_tokens: int
    max_output_tokens: int
    max_cost_usd: float


@dataclass(frozen=True)
class ParseResult:
    document: dict[str, Any]
    usage: ProviderUsage
    adapter_trace_id: str


@dataclass(frozen=True)
class OCRRequest:
    source_sha256: str
    max_pages: int


@dataclass(frozen=True)
class OCRPage:
    page_number: int
    text: str
    confidence: float
