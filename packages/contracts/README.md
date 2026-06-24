# Roliq contracts

This package is the single source of truth for public API, event, artifact-pointer, and Resume Document schemas.

Canonical Phase 2 contracts:

- `resume/resume-document.v1.schema.json`
- `common/artifact-pointer.v1.schema.json`
- `events/resume.parse.requested.v1.json`
- `events/resume.parse.completed.v1.json`
- `events/resume.parse.failed.v1.json`

Go, Python, and TypeScript validators load these files directly. Do not copy schemas into a provider adapter or language package.

## Validate

```powershell
go test ./packages/contracts
pnpm --filter @roliq/contracts test
$env:PYTHONPATH="$PWD\services\ai"
python -m pytest services/ai/tests
```

Parsing events are pointer-only. Adding resume text, contact data, provider payloads, prompts, or OCR text to an event is a contract violation.
