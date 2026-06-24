# ADR 0002: Evidence-backed Resume Document and immutable versions

**Status:** Accepted

## Context

Phase 2 must turn an untrusted resume into structured data without inventing facts or silently changing the user's career profile. The structured representation will be consumed by the editor, later matching, and later document-generation phases, so an unversioned provider-specific payload would become a long-lived product liability.

## Decision

Roliq will define a provider-neutral `resume-document.v1` JSON Schema using JSON Schema draft 2020-12. The schema is the canonical boundary shared by Go, Python, and TypeScript.

The document will:

- model candidate basics, summary, skills, experience, education, projects, certifications, languages, and links;
- distinguish extracted facts from normalized display values;
- attach source evidence to parsed facts using document location plus a bounded supporting excerpt;
- carry confidence, warnings, and review state without treating confidence as truth;
- prohibit unknown fields at contract boundaries;
- identify its schema version and source resume version;
- contain no provider-specific response objects or chain-of-thought content.

`resume_versions` remains the immutable version ledger:

- the Phase 1 original version references the security-cleared file;
- a successful parser creates one child version with source `parsed`;
- user corrections create a child version with source `manual` rather than mutating the parsed version;
- the selected active version is an explicit pointer, not an in-place rewrite;
- profile merge is a separate audited operation and never follows automatically from parsing.

Validated structured JSON may be stored transactionally with the resume version. Extracted text, OCR output, and provider request/response artifacts remain private checksummed object-store artifacts with database metadata and retention controls.

## Consequences

- Provider adapters must map into the Roliq schema and cannot leak their native response shape into product code.
- Schema evolution requires a new version and explicit migration/read compatibility; existing version payloads are immutable.
- Evidence increases payload size but makes review, correction, quality evaluation, and later truthful generation possible.
- The UI must preserve unsupported or low-confidence fields visibly instead of silently dropping or accepting them.

## Rejected alternatives

- Using a provider response as the canonical model couples the product to one vendor and makes validation inconsistent.
- Overwriting one mutable resume JSON document destroys auditability and makes concurrent editing unsafe.
- Writing parsed fields directly into the career profile violates review-first product policy.
