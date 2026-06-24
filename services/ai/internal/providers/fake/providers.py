from __future__ import annotations

import json
import os
from pathlib import Path
from typing import Any

from internal.contracts import validate_resume_document
from internal.providers.interfaces import (
    OCRPage,
    OCRRequest,
    ParseRequest,
    ParseResult,
    ProviderError,
    ProviderErrorCategory,
    ProviderUsage,
)


def _fixtures_root() -> Path:
    configured = os.getenv("ROLIQ_RESUME_FIXTURES_DIR")
    if configured:
        root = Path(configured).resolve()
        if not root.is_dir():
            raise RuntimeError(f"ROLIQ_RESUME_FIXTURES_DIR does not exist: {root}")
        return root
    for parent in Path(__file__).resolve().parents:
        candidate = parent / "fixtures" / "resumes"
        if candidate.is_dir():
            return candidate
    raise RuntimeError("Could not locate fixtures/resumes; set ROLIQ_RESUME_FIXTURES_DIR")


class _FixtureRepository:
    def __init__(self, root: Path | None = None):
        self.root = (root or _fixtures_root()).resolve()
        manifest = self._read_json(self.root / "manifest.json")
        if manifest.get("syntheticOnly") is not True:
            raise RuntimeError("Fake providers require a synthetic-only fixture manifest")
        self._by_checksum = {
            fixture["sourceSha256"]: fixture for fixture in manifest.get("fixtures", [])
        }

    def fixture(self, checksum: str) -> dict[str, Any]:
        try:
            return self._by_checksum[checksum]
        except KeyError as error:
            raise ProviderError(
                ProviderErrorCategory.PERMANENT_INPUT, "unknown_synthetic_fixture", False
            ) from error

    def expected_document(self, fixture: dict[str, Any]) -> dict[str, Any]:
        path = self.root / fixture["directory"] / fixture["expectedDocument"]
        return self._read_json(path)

    def ocr_output(self, fixture: dict[str, Any]) -> dict[str, Any]:
        path = self.root / fixture["directory"] / fixture["fakeOcrOutput"]
        return self._read_json(path)

    @staticmethod
    def _read_json(path: Path) -> dict[str, Any]:
        with path.open("r", encoding="utf-8") as handle:
            value = json.load(handle)
        if not isinstance(value, dict):
            raise RuntimeError(f"Fixture JSON must be an object: {path}")
        return value


class FakeResumeParserProvider:
    """Deterministically maps a synthetic source checksum to its golden document."""

    def __init__(self, fixture_root: Path | None = None):
        self._fixtures = _FixtureRepository(fixture_root)

    def parse(self, request: ParseRequest) -> ParseResult:
        if request.schema_version != "resume-document.v1":
            raise ProviderError(ProviderErrorCategory.VALIDATION, "unsupported_schema", False)
        if request.max_input_tokens <= 0 or request.max_output_tokens <= 0 or request.max_cost_usd < 0:
            raise ProviderError(ProviderErrorCategory.BUDGET_EXCEEDED, "invalid_budget", False)
        fixture = self._fixtures.fixture(request.source_sha256)
        document = self._fixtures.expected_document(fixture)
        validate_resume_document(document)
        return ParseResult(
            document=document,
            usage=ProviderUsage(),
            adapter_trace_id=f"fake:{fixture['id']}:{request.source_sha256[:12]}",
        )


class FakeOCRProvider:
    """Returns committed synthetic page text without performing OCR or I/O outside fixtures."""

    def __init__(self, fixture_root: Path | None = None):
        self._fixtures = _FixtureRepository(fixture_root)

    def extract(self, request: OCRRequest) -> tuple[OCRPage, ...]:
        if request.max_pages <= 0:
            raise ProviderError(ProviderErrorCategory.BUDGET_EXCEEDED, "ocr_page_budget", False)
        fixture = self._fixtures.fixture(request.source_sha256)
        output = self._fixtures.ocr_output(fixture)
        raw_pages = output.get("pages", [])
        if len(raw_pages) > request.max_pages:
            raise ProviderError(ProviderErrorCategory.BUDGET_EXCEEDED, "ocr_page_budget", False)
        return tuple(
            OCRPage(
                page_number=int(page["pageNumber"]),
                text=str(page["text"]),
                confidence=float(page["confidence"]),
            )
            for page in raw_pages
        )
