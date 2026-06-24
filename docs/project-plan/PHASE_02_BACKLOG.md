# Phase 2 Delivery Backlog

Status: **PLANNING ACTIVE - RUNTIME NOT STARTED**

Task IDs are stable references for progress updates and pull requests. A task is complete only when its acceptance evidence is recorded in `PROGRESS.md`.

## Sequencing rules

- Contracts precede migrations, handlers, workers, and UI.
- Deterministic extraction precedes OCR and LLM parsing.
- Synthetic fixtures and fake adapters precede real-provider calls.
- Consent and retention policy must be enforced before any real document reaches an AI/OCR provider.
- Phase 1 interactive browser closeout may run in parallel with Phase 2 planning/contracts, but Phase 2 runtime starts only after that gate closes.

## Milestone P2-M0 - Decisions and contracts

| ID     | Status      | Task                                              | Depends on               | Acceptance evidence                                                                           |
| ------ | ----------- | ------------------------------------------------- | ------------------------ | --------------------------------------------------------------------------------------------- |
| P2-001 | NOT_STARTED | Resolve LLM provider/model and accept its ADR     | None                     | P2-D01 closed with official terms, evaluation plan, cost ceiling, and accepted ADR            |
| P2-002 | NOT_STARTED | Resolve OCR adapter and accept its ADR            | None                     | P2-D02 closed with region, retention, provenance, limits, and accepted ADR                    |
| P2-003 | NOT_STARTED | Resolve workload identity/service auth            | None                     | Least-privilege policy and local mechanism documented in accepted ADR                         |
| P2-004 | NOT_STARTED | Approve consent and artifact retention policy     | None                     | Versioned consent copy, retention/deletion matrix, and accepted ADR                           |
| P2-005 | COMPLETE    | Define parsing evaluation framework               | ADR 0002                 | Golden strategy, measurable metrics, regression policy, and methodology accepted in ADR 0004  |
| P2-006 | COMPLETE    | Define provider isolation layer                   | ADR 0002, ADR 0003       | Roliq-owned interfaces and adapter boundaries accepted in ADR 0005                            |
| P2-007 | COMPLETE    | Define AI cost governance                         | ADR 0003                 | Document, retry, token, cost, cancellation, duplicate, and queue limits accepted in ADR 0006  |
| P2-010 | IN_REVIEW   | Define `resume-document.v1.schema.json`           | ADR 0002, ADR 0004       | Draft 2020-12 schema, valid/invalid fixtures, evidence/confidence rules, schema tests         |
| P2-011 | IN_REVIEW   | Define requested/completed/failed parsing events  | ADR 0003, P2-010         | Pointer-only JSON Schemas with cross-language and PII-rejection tests                         |
| P2-012 | IN_REVIEW   | Implement shared cross-language validation layer  | P2-010, P2-011           | Go, Python, and TypeScript validate canonical files with no schema copies                     |
| P2-013 | IN_REVIEW   | Build synthetic resume fixture and golden catalog | ADR 0004, P2-010         | Six original PDFs, expected documents, notes, metadata, fake OCR, and validation tooling      |
| P2-014 | NOT_STARTED | Design Phase 2 OpenAPI surface                    | P2-010                   | Reviewed operations for jobs, versions, corrections, active selection, merge, and preferences |
| P2-015 | IN_REVIEW   | Add provider interfaces and deterministic fakes   | ADR 0005, P2-010, P2-013 | Offline fake parser/OCR contract tests return exact fixture results and zero external cost    |

Exit: contracts are reviewable and vendor-independent; open decisions clearly block only their affected slices.

## Milestone P2-M1 - Persistence and orchestration

| ID     | Task                                    | Depends on             | Acceptance evidence                                                                                                                   |
| ------ | --------------------------------------- | ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| P2-020 | Add append-only Phase 2 migration       | P2-010, P2-011, P2-004 | Add `parsed` resume source plus jobs, attempts, artifacts, active version, merge audit, consent, preferences, indexes, RLS, down test |
| P2-021 | Add sqlc queries and repository methods | P2-020                 | Generated code compiles; tenant and optimistic-conflict integration tests pass                                                        |
| P2-022 | Implement parsing command/state machine | P2-011, P2-021         | Idempotent request/retry/cancel, compare-and-set transitions, stale-job recovery tests                                                |
| P2-023 | Publish parsing requests through outbox | P2-022                 | Pointer-only SQS event, retry/DLQ behavior, duplicate delivery tests                                                                  |
| P2-024 | Consume validated completion artifacts  | P2-010, P2-022         | Checksum/schema/tenant validation creates exactly one immutable parsed version                                                        |

Exit: a fake completion event can move a ready resume to `needs_review` without Python or a real provider.

## Milestone P2-M2 - Deterministic AI-service foundation

| ID     | Task                                              | Depends on             | Acceptance evidence                                                                                              |
| ------ | ------------------------------------------------- | ---------------------- | ---------------------------------------------------------------------------------------------------------------- |
| P2-030 | Scaffold `services/ai` FastAPI service and worker | P2-003                 | Locked dependencies, config validation, health/readiness, structured logs, metrics, traces, container/Compose/CI |
| P2-031 | Implement secure artifact client                  | P2-003, P2-030         | Scoped prefixes, checksum verification, size/decompression limits, no PII logs                                   |
| P2-032 | Implement PDF extraction with provenance          | P2-013, P2-031         | Golden fixtures pass for pages, columns, tables, corrupt, encrypted, and textless inputs                         |
| P2-033 | Implement DOCX extraction with provenance         | P2-013, P2-031         | Golden paragraph/table provenance and zip-bomb/relationship safety tests pass                                    |
| P2-034 | Implement OCR-needed classification               | P2-032                 | Deterministic threshold and visible `needs_ocr` state; no provider call yet                                      |
| P2-035 | Implement configured OCR adapter                  | P2-002, P2-004, P2-034 | Fake and production adapter contract tests, timeout/retry/cost metrics                                           |

Exit: security-cleared fixtures produce checksummed extraction artifacts or classified actionable failures without an LLM.

## Milestone P2-M3 - Schema-constrained parsing

| ID     | Task                                             | Depends on             | Acceptance evidence                                                               |
| ------ | ------------------------------------------------ | ---------------------- | --------------------------------------------------------------------------------- |
| P2-040 | Define LLM adapter and deterministic fake        | P2-010, P2-030         | Timeout, token/cost ceiling, structured output, redaction, and fake-adapter tests |
| P2-041 | Implement prompt/data boundary and normalization | P2-013, P2-040         | Malicious instructions remain data; unsupported facts are rejected or warned      |
| P2-042 | Implement selected provider adapter              | P2-001, P2-004, P2-041 | Schema reliability evaluation and provider failure tests pass                     |
| P2-043 | Publish completion/failure pointer events        | P2-011, P2-031, P2-042 | No PII in messages; checksum, idempotency, retry, and cancellation tests pass     |
| P2-044 | Complete end-to-end fake-provider pipeline       | P2-024, P2-043         | Ready resume reaches `needs_review`; duplicates create no extra version or charge |

Exit: the async pipeline works end to end with the deterministic fake and is safe to enable for an approved provider.

## Milestone P2-M4 - Product API and review experience

| ID     | Task                                         | Depends on     | Acceptance evidence                                                                     |
| ------ | -------------------------------------------- | -------------- | --------------------------------------------------------------------------------------- |
| P2-050 | Implement parsing/version API                | P2-014, P2-024 | Generated client and authorization, validation, idempotency, pagination, conflict tests |
| P2-051 | Implement corrected manual versions          | P2-050         | Immutable child version, optimistic concurrency, reload and audit tests                 |
| P2-052 | Implement explicit profile merge             | P2-051         | Field selection, diff, audit, conflict, and no-silent-overwrite tests                   |
| P2-053 | Implement career preferences persistence/API | P2-014, P2-020 | Global location/currency/time-zone validation, versioning, RLS tests                    |
| P2-054 | Build parsing status and failure UI          | P2-050         | Honest queued/OCR/failure/retry/cancel states; no fake progress                         |
| P2-055 | Build evidence-backed resume editor/history  | P2-051         | Keyboard-accessible review, confidence/warning display, local draft, version labels     |
| P2-056 | Build profile merge review UI                | P2-052, P2-055 | Explicit selected fields and persisted audit confirmation                               |
| P2-057 | Build job-preferences UI                     | P2-053         | Complete validated preference model survives reload                                     |

Exit: a user can review, correct, version, merge selected facts, and persist preferences without silent mutation.

## Milestone P2-M5 - Hardening and closeout

| ID     | Task                                                          | Depends on    | Acceptance evidence                                                                        |
| ------ | ------------------------------------------------------------- | ------------- | ------------------------------------------------------------------------------------------ |
| P2-060 | Add provider/queue/artifact operational metrics and alerts    | P2-M3         | Cost, latency, retries, OCR, schema rejection, and correction-rate dashboards              |
| P2-061 | Add parsing, outage, consent-revocation, and cleanup runbooks | P2-004, P2-M3 | Failure drills have commands, ownership, rollback, and safe recovery                       |
| P2-062 | Complete security and privacy review                          | P2-M4         | Threat model, PII log scan, least privilege, retention/deletion, dependency/container scan |
| P2-063 | Complete integration/browser/resilience acceptance            | P2-M4, P2-060 | Phase 2 exit criteria pass from clean Compose state and evidence is recorded               |

Exit: every criterion in `PHASE_02_PLAN.md` has current evidence and no phase-blocking issue remains.

## First implementation slice after planning

Finish verification of P2-010 through P2-013 and P2-015 before beginning P2-014. Do not begin persistence, FastAPI, queue orchestration, extraction, OCR, or real-provider work as part of this hardening slice.
