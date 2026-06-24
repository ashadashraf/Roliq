# Phase 2 Decision Register

Status: **IN_PROGRESS**

This register separates architecture already accepted from vendor, legal, and policy choices that must be resolved before the affected runtime slice begins.

## Accepted architecture

| Decision | Outcome                                                                   | Evidence                        |
| -------- | ------------------------------------------------------------------------- | ------------------------------- |
| P2-A01   | Resume data uses a provider-neutral evidence-backed versioned contract    | ADR 0002                        |
| P2-A02   | Parsed and manual resume versions are immutable                           | ADR 0002                        |
| P2-A03   | Parsing is asynchronous; Go owns jobs and Python has no product DB access | ADR 0003                        |
| P2-A04   | Queue events contain pointers/checksums, not resume text                  | ADR 0003                        |
| P2-A05   | Profile merge is explicit, field-level, and audited                       | `PHASE_02_PLAN.md` and ADR 0002 |
| P2-A06   | Parsing quality changes require a synthetic golden evaluation gate        | ADR 0004                        |
| P2-A07   | Provider-native types remain inside adapters; canonical output is v1      | ADR 0005                        |
| P2-A08   | Every logical parse has hard document, retry, token, and cost ceilings    | ADR 0006                        |

## Open blocking decisions

### P2-D01 - LLM provider and model

- Status: OPEN
- Blocks: provider adapter and real-provider integration; does not block schema, fixtures, deterministic extraction, or fake-adapter contract tests.
- Required evidence: structured-output reliability, context limits, supported regions, retention/training terms, deletion controls, rate limits, price ceilings, outage behavior, and an evaluation on synthetic resumes.
- Recommended decision shape: one primary adapter, one deterministic fake adapter for tests, and a provider-neutral interface. Do not implement automatic fallback until cost and duplicate-processing semantics are defined.
- Owner: product/security/engineering.

### P2-D02 - OCR production adapter

- Status: OPEN
- Blocks: production OCR path only.
- Required evidence: supported PDF/image formats, regional processing, retention terms, table/layout provenance, asynchronous limits, cost, throttling, and confidence semantics.
- Recommended default: AWS Textract behind an OCR interface because AWS is the reference deployment; keep a deterministic fake adapter for local and CI use.
- Owner: platform/security/engineering.

### P2-D03 - AI workload identity and local service authentication

- Status: OPEN
- Blocks: deployment of the AI worker with real object-store/queue access.
- Required evidence: least-privilege read/write prefixes, queue permissions, credential rotation, local-development parity, token audience/expiry, and incident revocation.
- Recommended default: cloud workload identity/IAM in production; non-production static LocalStack credentials plus short-lived audience-bound internal tokens where HTTP service calls exist.
- Owner: platform/security.

### P2-D04 - Consent, artifact retention, and deletion policy

- Status: OPEN
- Blocks: persistence migration and any real AI-provider call.
- Required product decisions: consent text/version, whether parsing is opt-in per user or per resume, extracted-text retention, provider artifact retention, consent revocation behavior, account deletion behavior, and audit retention.
- Engineering default until approved: no real provider call; synthetic fixtures only; private artifacts with explicit expiry metadata.
- Owner: product/legal/security.

## Decision protocol

For each open decision:

1. Record current official provider/legal evidence and the review date.
2. Compare at least the recommended option and one viable alternative.
3. Document security, privacy, cost, operational, and lock-in consequences.
4. Create a dedicated ADR and mark it Accepted before merging the blocked runtime slice.
5. Put secrets and commercial terms in approved secret/configuration systems, never in the repository.
