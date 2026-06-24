# Phase 2 Implementation Plan

Status: **IN_PROGRESS - PLANNING ONLY**

Runtime implementation has not started. The dependency-ordered delivery backlog is in `PHASE_02_BACKLOG.md`; unresolved vendor, identity, consent, and retention choices are tracked in `PHASE_02_DECISIONS.md`.

Accepted architecture:

- ADR 0002 defines the evidence-backed Resume Document and immutable version model.
- ADR 0003 defines asynchronous parsing with pointer-only events and no Python access to product tables.
- ADR 0004 defines the synthetic golden dataset, measurable quality metrics, and regression policy.
- ADR 0005 isolates provider-native schemas and SDKs behind Roliq-owned interfaces.
- ADR 0006 defines document, retry, token, cost, cancellation, duplicate, and queue safeguards.

## Goal

Convert a security-cleared resume into structured, evidence-backed data that the user can review, correct, version, and explicitly merge into their career profile. Add complete job preferences. Do not ingest jobs, generate application documents, or apply anywhere.

## Required decisions before runtime integration

Record each choice as an ADR and close the corresponding decision-register item:

1. P2-D01: initial LLM provider and model, including data-retention and regional-processing terms.
2. P2-D02: production OCR adapter for textless PDFs; recommended default is an interface with AWS Textract as the reference adapter.
3. P2-D03: AI service identity mechanism; recommended default is workload identity/IAM in AWS and signed short-lived service tokens locally.
4. P2-D04: candidate consent copy and retention period for extracted text and AI artifacts.

Defaults already fixed:

- Parsing is review-first and never overwrites the career profile.
- Deterministic extraction precedes any LLM call.
- AI providers sit behind an internal interface.
- Results include schema version, source evidence, confidence, warnings, and provider trace metadata without storing prompts containing PII in logs.
- Original and parsed versions are immutable; corrections create a new manual version.

## Target data flow

1. The Phase 1 worker marks a resume `ready` and publishes `resume.scan.completed.v1`.
2. The Go orchestration layer idempotently creates a parsing job and publishes a pointer-only queue message.
3. The Python consumer downloads the clean object using scoped credentials, verifies checksum, and extracts text deterministically.
4. Textless documents enter `needs_ocr`; configured OCR produces a separate checksummed artifact.
5. The parser sends untrusted resume text to a schema-constrained LLM adapter as data, validates the response with Pydantic and JSON Schema, and rejects unsupported claims.
6. The AI service stores the structured result as a private, checksummed artifact and publishes a completion pointer; resume PII is not placed directly in SQS.
7. The Go worker validates the artifact, creates an immutable parsed resume version, and changes the job to `needs_review`.
8. The user reviews and corrects fields. Saving creates a manual version with optimistic concurrency.
9. The user explicitly selects fields to merge into the career profile.

## Workstreams and order

The stable task order and task-level acceptance evidence are maintained in `PHASE_02_BACKLOG.md`. The sections below define scope and design intent.

### 1. Contracts and schema

- Define `resume-document.v1.json` with basics, summary, skills, experience, education, projects, certifications, languages, links, and per-field evidence.
- Define `resume.parsing.requested.v1`, `completed.v1`, and `failed.v1` pointer-only event schemas.
- Add parsing states: `queued`, `extracting`, `needs_ocr`, `parsing`, `needs_review`, `completed`, `failed`, and `canceled`.
- Generate Python Pydantic and TypeScript types from the versioned contract where practical.

### 2. Persistence and orchestration

- Add `resume_parsing_jobs`, artifact references, attempt history, active resume version, and profile-merge audit records.
- Use unique organization/resume/source-version constraints for idempotency.
- Add queue publishing and completion consumption through the existing outbox pattern.
- Add retry limits, exponential backoff, dead-letter state, cancellation, and stale-job recovery.

### 3. AI service foundation

- Scaffold FastAPI health/readiness endpoints, configuration validation, JSON logs, OpenTelemetry, metrics, and service authentication.
- Add PDF and DOCX extraction adapters with page/paragraph provenance.
- Detect encrypted, corrupt, oversized-after-expansion, and textless documents.
- Add OCR and LLM interfaces plus one configured adapter for each approved dependency.
- Enforce timeouts, token/cost ceilings, schema validation, redacted error reporting, and prompt-injection-safe templates.

### 4. Product API

- Add endpoints to request/retry/cancel parsing, read job status, list versions, read a structured version, create a corrected version, select the active version, compare version metadata, and merge approved fields.
- Add career-preference GET/PUT endpoints with optimistic versioning.
- Extend OpenAPI first and regenerate clients before UI work.

### 5. Web experience

- Show honest parsing states and actionable failures on resume pages and dashboard.
- Build a sectioned resume editor with evidence/confidence warnings and autosaved local drafts.
- Add version history and clear original/parsed/manual labels.
- Add field-level profile merge review; no one-click silent overwrite.
- Add preferences for target roles, skills, seniority, job types, remote modes, locations, work authorization, salary/currency/period, time zones/hours, industries, company size, travel, perks, and exclusions.

### 6. Security and operations

- Capture explicit AI-processing consent with policy version and timestamp.
- Keep raw text and artifacts private, encrypted, tenant-scoped, and excluded from logs/traces.
- Treat document text as untrusted content and prohibit tool/instruction execution from it.
- Add cost, latency, failure, retry, OCR, schema-rejection, and review-correction metrics.
- Add runbooks for stuck jobs, provider outage, invalid output, consent revocation, and artifact cleanup.

### 7. Testing

- Use synthetic PDF/DOCX fixtures covering standard, multi-page, columns, tables, textless, corrupt, encrypted, and malicious-instruction documents.
- Golden-test deterministic extraction and schema normalization.
- Contract-test events and artifacts across Go/Python/TypeScript.
- Test malformed/truncated LLM output, hallucinated fields, low confidence, provider timeouts, duplicate events, retries, DLQ, and cancellation.
- Prove tenant isolation, version immutability, optimistic conflicts, explicit merge, and consent checks.
- Browser-test upload-to-review, correction, version history, merge, and preferences.

## Phase 2 exit criteria

- A valid ready resume reaches `needs_review` without synchronous request blocking.
- Every parsed field has supported provenance or is visibly marked unsupported/low-confidence.
- Corrections create an immutable manual version and survive reload.
- Profile changes occur only after explicit field approval and are audited.
- Preferences persist, validate globally aware location/currency/time-zone data, and are tenant isolated.
- Duplicate events and retries do not create duplicate versions or charges.
- Provider outage, invalid output, corrupt input, and OCR-needed states are recoverable and visible.
- Compose, CI, contract, integration, security, and browser suites pass.

## Explicit non-goals

- Job ingestion, search, ranking, or recommendations.
- Resume variations or cover letters.
- Application approval, tracking, or submission.
- Billing activation.
- Playwright automation.

## Planning completion gate

Planning is complete when:

- ADRs 0002 and 0003 have been reviewed;
- P2-D01 through P2-D04 have named owners and decision dates;
- the first contract slice P2-010/P2-011/P2-013 is approved to begin;
- Phase 1 interactive browser acceptance is either complete or explicitly waived by the product owner before Phase 2 runtime work begins.
