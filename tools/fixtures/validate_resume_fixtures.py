from __future__ import annotations

import hashlib
import json
import re
from pathlib import Path
from typing import Any, Iterable

from pypdf import PdfReader

from internal.contracts import validate_resume_document


ROOT = Path(__file__).resolve().parents[2]
FIXTURE_ROOT = ROOT / "fixtures" / "resumes"
EXPECTED_SCENARIOS = {
    "senior_backend_engineer",
    "frontend_engineer",
    "data_scientist",
    "fresh_graduate",
    "product_manager",
    "career_switcher",
}


def read_json(path: Path) -> dict[str, Any]:
    with path.open("r", encoding="utf-8") as handle:
        value = json.load(handle)
    if not isinstance(value, dict):
        raise AssertionError(f"Expected JSON object: {path}")
    return value


def checksum(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def normalized(value: str) -> str:
    return re.sub(r"\s+", " ", value).strip().casefold()


def evidence_items(value: Any) -> Iterable[dict[str, Any]]:
    if isinstance(value, dict):
        evidence = value.get("evidence")
        if isinstance(evidence, list):
            yield from (item for item in evidence if isinstance(item, dict))
        for child in value.values():
            yield from evidence_items(child)
    elif isinstance(value, list):
        for child in value:
            yield from evidence_items(child)


def validate_fixture(directory: Path, manifest_entry: dict[str, Any]) -> None:
    required = {
        "resume.pdf",
        "expected_resume_document_v1.json",
        "evaluation_notes.md",
        "evaluation_metadata.json",
        "fake_ocr_output.json",
    }
    missing = required - {path.name for path in directory.iterdir() if path.is_file()}
    if missing:
        raise AssertionError(f"{directory.name} missing files: {sorted(missing)}")

    pdf_path = directory / "resume.pdf"
    source_hash = checksum(pdf_path)
    if source_hash != manifest_entry["sourceSha256"]:
        raise AssertionError(f"{directory.name} manifest checksum mismatch")

    reader = PdfReader(str(pdf_path))
    metadata = read_json(directory / "evaluation_metadata.json")
    if metadata.get("synthetic") is not True:
        raise AssertionError(f"{directory.name} must be explicitly synthetic")
    if metadata.get("sourceSha256") != source_hash:
        raise AssertionError(f"{directory.name} metadata checksum mismatch")
    if metadata.get("pageCount") != len(reader.pages):
        raise AssertionError(f"{directory.name} page count mismatch")

    document = read_json(directory / "expected_resume_document_v1.json")
    validate_resume_document(document)
    if document["source"]["contentSha256"] != source_hash:
        raise AssertionError(f"{directory.name} document checksum mismatch")
    if document["source"]["pageCount"] != len(reader.pages):
        raise AssertionError(f"{directory.name} document page count mismatch")
    email = document.get("candidate", {}).get("email", {}).get("value", "")
    if not email.endswith("@example.com"):
        raise AssertionError(f"{directory.name} must use an example.com address")

    page_text = [normalized(page.extract_text() or "") for page in reader.pages]
    for evidence in evidence_items(document):
        if evidence["sourceSha256"] != source_hash:
            raise AssertionError(f"{directory.name} evidence checksum mismatch")
        page_number = evidence.get("page")
        if page_number is None or not 1 <= page_number <= len(page_text):
            raise AssertionError(f"{directory.name} evidence page is invalid")
        if normalized(evidence["excerpt"]) not in page_text[page_number - 1]:
            raise AssertionError(
                f"{directory.name} evidence not found on page {page_number}: {evidence['excerpt']!r}"
            )

    fake_ocr = read_json(directory / "fake_ocr_output.json")
    if fake_ocr.get("sourceSha256") != source_hash:
        raise AssertionError(f"{directory.name} fake OCR checksum mismatch")
    if len(fake_ocr.get("pages", [])) != len(reader.pages):
        raise AssertionError(f"{directory.name} fake OCR page count mismatch")

    notes = (directory / "evaluation_notes.md").read_text(encoding="utf-8")
    if "entirely synthetic" not in notes.casefold():
        raise AssertionError(f"{directory.name} notes must declare synthetic provenance")


def main() -> None:
    manifest = read_json(FIXTURE_ROOT / "manifest.json")
    if manifest.get("syntheticOnly") is not True:
        raise AssertionError("Fixture manifest must be synthetic-only")
    entries = {item["id"]: item for item in manifest.get("fixtures", [])}
    if set(entries) != EXPECTED_SCENARIOS:
        raise AssertionError(
            f"Fixture scenarios differ: expected {sorted(EXPECTED_SCENARIOS)}, got {sorted(entries)}"
        )
    for scenario in sorted(entries):
        validate_fixture(FIXTURE_ROOT / entries[scenario]["directory"], entries[scenario])
        print(f"PASS {scenario}")


if __name__ == "__main__":
    main()
