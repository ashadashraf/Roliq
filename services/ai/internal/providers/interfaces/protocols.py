from __future__ import annotations

from typing import Protocol

from .models import OCRPage, OCRRequest, ParseRequest, ParseResult


class ResumeParserProvider(Protocol):
    def parse(self, request: ParseRequest) -> ParseResult: ...


class OCRProvider(Protocol):
    def extract(self, request: OCRRequest) -> tuple[OCRPage, ...]: ...
