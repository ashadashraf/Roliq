# ADR 0006: AI cost governance

**Status:** Accepted

## Context

Resume extraction, OCR, and LLM parsing consume bounded but non-zero compute and vendor spend. Duplicate delivery, pathological documents, retries, or unrestricted concurrency could create unexpected cost before billing or entitlement enforcement exists.

## Decision

Every parsing job carries an immutable budget snapshot. Work is rejected or stopped when a hard limit is reached; limits cannot be raised implicitly by retries.

### Resume limits

| Control                          | Foundation default                                                 |
| -------------------------------- | ------------------------------------------------------------------ |
| Uploaded file size               | 10 MiB, inherited from Phase 1                                     |
| Maximum document pages           | 25 pages                                                           |
| Maximum OCR pages                | 10 pages; pages beyond the limit require explicit user remediation |
| Maximum parsing attempts         | 3 total attempts, including the initial attempt                    |
| Maximum extracted text           | 250,000 Unicode characters before normalization                    |
| Maximum expanded archive content | 50 MiB for DOCX safety                                             |

Limits are configuration with validated upper bounds, not scattered constants. A lower product-plan limit may be applied without changing stored historical jobs.

### Per-parse budgets

- Default maximum LLM input: 50,000 tokens.
- Default maximum LLM output: 8,000 tokens.
- Default maximum estimated provider spend: USD 0.25 per logical parse job across all attempts.
- OCR and LLM usage are reserved before dispatch and reconciled after completion.
- A retry receives only the unspent remainder of the original job budget.
- A provider call is not dispatched when price metadata is missing, stale, or would exceed the remaining budget.

Provider prices remain configuration with source and effective date; they are not embedded in business logic. Budget enforcement uses a pessimistic estimate before dispatch and records provider-neutral usage afterward.

### Cancellation and duplicate protection

- Queued work cancels immediately.
- In-flight cancellation is best effort; late results for a canceled job are validated but not promoted to a resume version.
- One organization/source-version/schema/prompt combination has one logical active job.
- Queue delivery and completion are idempotent. A successful job creates at most one parsed version and consumes one logical entitlement.
- Retries require a classified retryable error and exponential backoff; permanent input and schema failures do not retry automatically.

### Queue safeguards

- Per-organization and global concurrency limits prevent one tenant from exhausting workers.
- Messages have bounded receive count and enter a DLQ after the attempt limit.
- Workers stop claiming expensive work when the provider circuit is open or budget/entitlement state cannot be read.
- Queue depth, oldest-message age, attempts, cancellations, estimated/actual cost, and budget rejections are observable.

### Future SaaS controls

Entitlement policy is separate from billing implementation. Initial planning targets are:

| Plan | Monthly included logical parses     |
| ---- | ----------------------------------- |
| Free | 5                                   |
| Pro  | 50                                  |
| Team | Configurable organization allowance |

These values are product configuration, not promises or active billing behavior. Phase 2 stores usage and enforces configured development entitlements; billing activation remains Phase 6.

## Consequences

- Some long resumes require user remediation instead of silently incurring unbounded cost.
- Budget snapshots improve auditability and make retries predictable.
- Accurate preflight estimates require versioned price metadata and tokenizer/counting adapters once a provider is selected.
- Fake adapters report deterministic zero monetary cost while still exercising limits and duplicate protection.
