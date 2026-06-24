# ADR 0003: Asynchronous resume parsing with pointer-only events

**Status:** Accepted

## Context

Extraction, OCR, and LLM parsing are slow, failure-prone, cost-bearing operations over sensitive, untrusted documents. Running them inside an API request or giving the Python service direct access to product tables would weaken reliability, tenant controls, and auditability.

## Decision

The Go product API owns parsing commands, job state, version creation, consent enforcement, and audit records. The Python AI runtime performs extraction and parsing but has no direct PostgreSQL access.

The workflow is asynchronous:

1. A command transaction creates or reuses a parsing job and writes a transactional outbox event.
2. A pointer-only queue message identifies organization, resume, source version, job, artifact location, checksums, schema version, and correlation metadata. Resume text and PII are not copied into queue bodies.
3. The AI worker obtains the clean source object through scoped object-store credentials, verifies its checksum, and writes checksummed extraction/result artifacts to a private prefix.
4. A pointer-only completion or failure event is published.
5. A trusted Go consumer validates job state, tenant identifiers, artifact checksum, and schema before atomically creating an immutable parsed version and moving the job to review.

The job state machine is `queued -> extracting -> needs_ocr -> parsing -> needs_review -> completed`, with terminal `failed` and `canceled` states. Transitions are compare-and-set operations. Duplicate messages are safe, one logical source version has at most one active parsing job, and one successful job creates at most one parsed version.

Retries use bounded exponential backoff with classified retryability. Exhausted jobs enter a visible dead-letter/failure state; they are never retried indefinitely. Cancellation prevents new expensive work but does not pretend to revoke an already-sent provider request.

## Consequences

- The UI polls or subscribes to durable job state and never waits on parsing synchronously.
- Artifact retention, queue policy, and workload identity are part of the security boundary.
- Cross-language contract tests are mandatory for every message and artifact schema.
- More orchestration code is required, but provider outages and duplicate delivery do not corrupt product state or duplicate charges.

## Rejected alternatives

- Synchronous parsing couples user requests to provider latency and makes retries unsafe.
- Direct AI-service database writes bypass product transactions, RLS conventions, and audit policy.
- Putting extracted text in SQS increases PII exposure, message size, and retention ambiguity.
