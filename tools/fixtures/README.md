# Synthetic resume fixtures

The fixture corpus is synthetic-only and governed by ADR 0004.

Generate deterministic PDFs and expected artifacts:

```powershell
python -m pip install -r tools/fixtures/requirements.txt
python tools/fixtures/generate_resume_fixtures.py
python tools/fixtures/render_resume_fixtures.py
```

Validate source checksums, PDF page counts, the canonical schema, evidence excerpts, metadata, and fake OCR output:

```powershell
$env:PYTHONPATH="$PWD\services\ai"
python tools/fixtures/validate_resume_fixtures.py
```

Generation is deterministic: rerunning it should not change committed fixture hashes or expected JSON. Never add a real resume, deliverable email address, copied resume template, or production-derived text to this directory.
