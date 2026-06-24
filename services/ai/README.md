# Roliq AI foundation

Phase 2 currently contains contracts, provider-owned interfaces, and deterministic offline fake adapters only. There is no FastAPI application, queue consumer, runtime parser, OCR integration, LLM integration, or external AI call.

Provider directories are deliberate extension boundaries governed by ADR 0005. Do not install a provider SDK or add network behavior until P2-D01/P2-D02 and the provider-specific ADR are accepted.

## Tests

```powershell
python -m pip install -e ".\services\ai[dev]"
$env:PYTHONPATH="$PWD\services\ai"
python -m pytest services/ai/tests
python tools/fixtures/validate_resume_fixtures.py
```

The fake parser maps a committed synthetic source checksum to its expected Resume Document. The fake OCR adapter returns committed synthetic page text. Both are deterministic, offline, and zero-cost by design.
