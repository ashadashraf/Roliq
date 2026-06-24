# Roliq Master Plan

Last reviewed: 2026-06-24

## Product goal

Roliq is a multi-tenant career and job-application SaaS. It converts a candidate's factual career history into structured, user-controlled data; finds and ranks relevant jobs; prepares role-specific documents; and, only with explicit user configuration, assists with or automates applications.

The product must never invent candidate facts, silently submit applications, or treat uploaded documents as trusted instructions.

## Architecture boundaries

- `apps/web`: Next.js product UI and Clerk authentication adapter.
- `services/api`: Go/Echo product API and source of truth for tenant-owned business state.
- `services/ai`: Python/FastAPI AI boundary beginning in Phase 2.
- `workers/background`: trusted cross-tenant platform work, outbox delivery, and document security processing.
- `workers/automation`: isolated Playwright runtime beginning in Phase 5.
- PostgreSQL: transactional business state with `organization_id` scoping and RLS.
- Redis: rate limiting, caching, and short-lived coordination.
- S3-compatible storage: quarantined documents and later AI artifacts.
- SNS/SQS-compatible messaging: asynchronous domain events and workload queues.
- OpenSearch: job search and retrieval beginning in Phase 3.

The Go API remains a modular monolith until operational evidence justifies extraction. AI and browser automation remain separate because their dependencies, scaling, data handling, and risk profiles differ.

## Roadmap

| Phase | Outcome                                                                  | Status                            |
| ----- | ------------------------------------------------------------------------ | --------------------------------- |
| 1     | SaaS foundation, identity, onboarding, profile, secure resume intake     | IN_PROGRESS - browser QA closeout |
| 2     | Resume parsing, structured resume editor, preferences, versioning        | IN_PROGRESS - planning only       |
| 3     | Job ingestion, search, matching, scoring, recommendations                | NOT_STARTED                       |
| 4     | Resume variations, cover letters, approval workflow, application tracker | NOT_STARTED                       |
| 5     | Isolated auto-apply workers, notifications, administration               | NOT_STARTED                       |
| 6     | Billing activation, production infrastructure, compliance and operations | NOT_STARTED                       |

## Product-wide rules

- Every business record is scoped to an organization, including personal workspaces.
- OIDC `issuer + subject` is the external identity key; email is not an authorization key.
- Long-running work is asynchronous, retryable, idempotent, observable, and auditable.
- The original resume and each derived version are immutable; edits create new versions.
- AI output is a draft until reviewed. It cannot overwrite factual career data automatically.
- Resume and job content are untrusted data and cannot issue system instructions.
- Application submission defaults to user approval. Later auto-apply requires explicit, revocable consent and policy checks.
- No fake production records, testimonials, metrics, jobs, or application results.
- Public contracts are versioned and generated from source-controlled schemas.
- Deployment code must use portable interfaces even when AWS is the reference platform.

## Definition of done for every phase

A phase is complete only when:

1. All declared user flows work end to end with persistent data.
2. Migrations, API/event contracts, UI behavior, and operational documentation agree.
3. Unit, integration, contract, tenant-isolation, failure-path, and relevant browser tests pass.
4. Security, rate limits, idempotency, audit behavior, logs, metrics, and traces are verified.
5. Docker Compose starts from a clean state and the production build succeeds.
6. No later-phase behavior is represented by fake or misleading product flows.
7. `PROGRESS.md` contains dated evidence and no unresolved phase-blocking item.
