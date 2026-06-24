# Implementation Phases

## Phase 1 — SaaS foundation

Goal: deliver the secure, persisted foundation on which later intelligence and automation can safely operate.

Deliverables:

- Editorial-light landing, Clerk sign-in/sign-up, protected application shell, and responsive dashboard.
- Provider-neutral OIDC/JWT verification and idempotent session bootstrap.
- Personal organization provisioning with membership, free internal subscription record, and tenant context.
- Resumable onboarding with resume and manual-profile paths.
- Structured career profile with experience, education, and skills.
- Presigned PDF/DOCX upload, quarantine, metadata and checksum validation, content validation, malware scanning, and explicit states.
- PostgreSQL schema, RLS, audit logs, transactional outbox, idempotency keys, migrations, and typed contract sources.
- Background worker, Docker Compose, Kubernetes base, OpenTelemetry/Prometheus seams, CI, tests, and developer documentation.

Exit criteria: every item in `PHASE_01_CLOSEOUT.md` is checked with evidence.

## Phase 2 — Career intelligence foundation

Goal: transform a clean resume into reviewable structured data and durable career preferences without changing factual data silently.

Deliverables:

- Python/FastAPI AI service and queue consumer with health, telemetry, and service authentication.
- Deterministic PDF/DOCX text extraction, OCR-needed detection, schema-constrained LLM parsing, and evidence/confidence metadata.
- Versioned Resume Document JSON schema and versioned parsing event contracts.
- Parsing job orchestration with idempotency, retries, dead-letter handling, artifact checksums, and user-visible states.
- Review-first resume editor that creates immutable manual versions.
- Career-profile merge flow requiring explicit field-level approval.
- Complete job-preference model and UI.
- Resume version list, comparison metadata, active-version selection, and optimistic concurrency.

Exit criteria: the workflow `ready resume → parse → review → corrected version → explicit profile merge → preferences complete` passes end to end, including corrupt, textless, malformed-AI-output, retry, and tenant-isolation cases.

## Phase 3 — Job discovery and matching

Goal: ingest legally usable jobs, make them searchable, and produce explainable recommendations.

Deliverables:

- Versioned job-source connector interface, source policy registry, ingestion scheduler, deduplication, normalization, and expiry.
- OpenSearch indexing with PostgreSQL source-of-truth records.
- Search, facets, saved filters, exclusions, and pagination.
- Hybrid hard-filter, lexical, semantic, and preference scoring.
- Explainable match breakdown and recommendation feedback.
- Source health, ingestion lag, duplicate rate, and ranking-quality telemetry.

Exit criteria: at least one approved real source operates reliably; stale and duplicate jobs are controlled; recommendations are tenant-safe, explainable, and measured against a labeled evaluation set.

## Phase 4 — Documents and approval workflow

Goal: prepare truthful role-specific application materials and track user-approved applications.

Deliverables:

- Resume variation generation from approved facts only.
- Cover-letter generation with claim provenance and review.
- Company, role, industry, location, and language targeting.
- Application draft, approval, rejection, and cancellation state machine.
- Manual application recording and application tracker.
- Artifact version links, audit history, and notifications for approval requests.

Exit criteria: generated claims trace to approved profile facts; document review is complete; no application can enter submission without a valid approval policy.

## Phase 5 — Controlled automation

Goal: submit supported applications through isolated workers with strong user control and recovery.

Deliverables:

- Docker-isolated Playwright workers with per-run credentials and network/resource limits.
- Site adapters, capability registry, policy/terms checks, and graceful unsupported-site handling.
- Approval-required and explicit auto-apply policies with revocation.
- Queue leasing, retries, screenshots/evidence, failure classification, and manual takeover.
- Notifications and an operational admin dashboard.

Exit criteria: supported sites pass sandbox/canary tests; no policy bypass or CAPTCHA circumvention exists; every submission has user authorization, evidence, audit history, and deterministic recovery behavior.

## Phase 6 — Commercial and production operations

Goal: operate Roliq as a secure, billable, observable SaaS on AWS while retaining portability.

Deliverables:

- Stripe-compatible billing activation, entitlements, metering, invoices, webhooks, and plan changes.
- Terraform for VPC, EKS, RDS, ElastiCache, S3/CloudFront, ECR, ALB, Route53, SES, SNS/SQS/DLQs, Secrets Manager, CloudWatch, and OpenSearch.
- Backup/restore, disaster recovery, deployment promotion, rollback, and secret rotation.
- SLOs, alerts, runbooks, audit retention, data export/deletion, and compliance evidence.
- Load, resilience, penetration, dependency, container, and recovery testing.

Exit criteria: production deployment, billing reconciliation, restore drill, rollback drill, alerting, privacy workflows, and operational ownership are proven.
