from __future__ import annotations

import json
import os
from pathlib import Path
from typing import Any

from jsonschema import Draft202012Validator, FormatChecker
from jsonschema.exceptions import ValidationError
from referencing import Registry, Resource


class ContractValidationError(ValueError):
    """A value failed the canonical Roliq JSON Schema contract."""


def _contracts_root() -> Path:
    configured = os.getenv("ROLIQ_CONTRACTS_DIR")
    if configured:
        root = Path(configured).resolve()
        if not root.is_dir():
            raise RuntimeError(f"ROLIQ_CONTRACTS_DIR does not exist: {root}")
        return root

    for parent in Path(__file__).resolve().parents:
        candidate = parent / "packages" / "contracts"
        if candidate.is_dir():
            return candidate
    raise RuntimeError("Could not locate packages/contracts; set ROLIQ_CONTRACTS_DIR")


def _load(path: Path) -> dict[str, Any]:
    with path.open("r", encoding="utf-8") as handle:
        value = json.load(handle)
    if not isinstance(value, dict) or not value.get("$id"):
        raise RuntimeError(f"Contract must be an object with $id: {path}")
    return value


_ROOT = _contracts_root()
_ARTIFACT_POINTER = _load(_ROOT / "common" / "artifact-pointer.v1.schema.json")
_RESUME_DOCUMENT = _load(_ROOT / "resume" / "resume-document.v1.schema.json")
_EVENT_FILES = {
    "resume.upload.completed.v1": "resume.upload.completed.v1.json",
    "resume.parse.requested.v1": "resume.parse.requested.v1.json",
    "resume.parse.completed.v1": "resume.parse.completed.v1.json",
    "resume.parse.failed.v1": "resume.parse.failed.v1.json",
}
_EVENT_SCHEMAS = {name: _load(_ROOT / "events" / file_name) for name, file_name in _EVENT_FILES.items()}
_REGISTRY = Registry().with_resource(
    _ARTIFACT_POINTER["$id"], Resource.from_contents(_ARTIFACT_POINTER)
)
_FORMAT_CHECKER = FormatChecker()
_RESUME_VALIDATOR = Draft202012Validator(
    _RESUME_DOCUMENT, registry=_REGISTRY, format_checker=_FORMAT_CHECKER
)
_EVENT_VALIDATORS = {
    name: Draft202012Validator(schema, registry=_REGISTRY, format_checker=_FORMAT_CHECKER)
    for name, schema in _EVENT_SCHEMAS.items()
}


def _validate(validator: Draft202012Validator, value: Any, contract_name: str) -> None:
    errors = sorted(validator.iter_errors(value), key=lambda error: list(error.absolute_path))
    if not errors:
        return
    details = "; ".join(_format_error(error) for error in errors[:10])
    raise ContractValidationError(f"{contract_name} validation failed: {details}")


def _format_error(error: ValidationError) -> str:
    path = ".".join(str(part) for part in error.absolute_path) or "$"
    return f"{path}: {error.message}"


def validate_resume_document(value: Any) -> None:
    _validate(_RESUME_VALIDATOR, value, "resume-document.v1")


def validate_event(event_type: str, value: Any) -> None:
    try:
        validator = _EVENT_VALIDATORS[event_type]
    except KeyError as error:
        raise ContractValidationError(f"Unsupported event contract: {event_type}") from error
    _validate(validator, value, event_type)
