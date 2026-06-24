from __future__ import annotations

import json
from pathlib import Path

import pytest

from internal.contracts import ContractValidationError, validate_event, validate_resume_document
from internal.providers.fake import FakeOCRProvider, FakeResumeParserProvider
from internal.providers.interfaces import OCRRequest, ParseRequest, ProviderError


ROOT = Path(__file__).resolve().parents[3]
FIXTURES = ROOT / "fixtures" / "resumes"
EVENTS = ROOT / "packages" / "contracts" / "testdata" / "events"


def read_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def test_all_golden_documents_validate() -> None:
    paths = sorted(FIXTURES.glob("*/expected_resume_document_v1.json"))
    assert len(paths) == 6
    for path in paths:
        validate_resume_document(read_json(path))


def test_unknown_provider_shape_is_rejected() -> None:
    document = read_json(FIXTURES / "fresh_graduate" / "expected_resume_document_v1.json")
    document["provider_response"] = {"unsafe": True}
    with pytest.raises(ContractValidationError):
        validate_resume_document(document)


@pytest.mark.parametrize(
    ("event_type", "file_name"),
    [
        ("resume.parse.requested.v1", "resume.parse.requested.v1.valid.json"),
        ("resume.parse.completed.v1", "resume.parse.completed.v1.valid.json"),
        ("resume.parse.failed.v1", "resume.parse.failed.v1.valid.json"),
    ],
)
def test_pointer_only_events_validate(event_type: str, file_name: str) -> None:
    validate_event(event_type, read_json(EVENTS / file_name))


def test_event_with_resume_text_is_rejected() -> None:
    with pytest.raises(ContractValidationError):
        validate_event(
            "resume.parse.requested.v1",
            read_json(EVENTS / "resume.parse.requested.v1.invalid-pii.json"),
        )


def test_fake_parser_returns_exact_valid_golden_document() -> None:
    manifest = read_json(FIXTURES / "manifest.json")
    fixture = manifest["fixtures"][0]
    provider = FakeResumeParserProvider(FIXTURES)
    result = provider.parse(
        ParseRequest(
            source_sha256=fixture["sourceSha256"],
            schema_version="resume-document.v1",
            prompt_version="fake.v1",
            max_input_tokens=50_000,
            max_output_tokens=8_000,
            max_cost_usd=0.25,
        )
    )
    expected = read_json(FIXTURES / fixture["directory"] / fixture["expectedDocument"])
    assert result.document == expected
    assert result.usage.estimated_cost_usd == 0
    assert result.adapter_trace_id.startswith("fake:")


def test_fake_ocr_is_deterministic_and_budgeted() -> None:
    manifest = read_json(FIXTURES / "manifest.json")
    fixture = manifest["fixtures"][1]
    provider = FakeOCRProvider(FIXTURES)
    first = provider.extract(OCRRequest(source_sha256=fixture["sourceSha256"], max_pages=10))
    second = provider.extract(OCRRequest(source_sha256=fixture["sourceSha256"], max_pages=10))
    assert first == second
    assert first[0].page_number == 1
    with pytest.raises(ProviderError):
        provider.extract(OCRRequest(source_sha256=fixture["sourceSha256"], max_pages=0))
