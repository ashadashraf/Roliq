# Phase 1 Closeout

Status: **implementation and automated acceptance complete; interactive browser QA pending**

Last verified: 2026-06-24

## Implemented

- [x] Next.js landing, authentication pages, protected shell, onboarding, career profile, resumes, and persisted dashboard.
- [x] Clerk frontend adapter with provider-neutral issuer/JWKS/audience verification in Go.
- [x] Idempotent bootstrap creating user, personal organization, owner membership, free internal subscription, onboarding state, and audit/outbox events.
- [x] Organization-scoped PostgreSQL schema, composite tenant references, RLS policies, migration runner, audit log, outbox, and idempotency storage.
- [x] Manual career profile persistence for headline, summary, location, experience, education, and skills.
- [x] Presigned quarantine uploads for PDF/DOCX with size, MIME, metadata, SHA-256, signature, DOCX-structure, and ClamAV checks.
- [x] Background scan state machine, expired-upload handling, and outbox publishing foundation.
- [x] OpenAPI contract, generated TypeScript schema, event contract, Dockerfiles, Compose, Kubernetes base, CI, tests, documentation, logs, metrics, and traces.
- [x] Later-phase AI, matching, billing, generation, and auto-apply behavior remains absent.

## Automated verification evidence

Current on 2026-06-24:

- [x] `go test ./...`, `go vet ./...`, and repository-wide `gofmt` check.
- [x] `pnpm format:check`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, and production `pnpm build`.
- [x] OpenAPI/Redocly validation and Linux-container `sqlc compile`.
- [x] Compose configuration parsing and all API, worker, and web container image builds.
- [x] Disposable Compose project started from fresh named volumes; migration exited zero, all services became healthy/running, API readiness returned 200, and web returned 200.
- [x] Migration ledger contains version 1; `roliq_app` can log in without `BYPASSRLS`; `roliq_worker` uses the separate cross-tenant role.
- [x] Real Clerk development JWT-template issuance proved discovery/JWKS, `roliq-api` audience, verified-email claims, and session bootstrap.
- [x] Both onboarding paths persisted across independent requests through the reusable acceptance suite.
- [x] Valid PDF and DOCX reached `ready`; invalid signature, checksum mismatch, and EICAR reached `rejected`; expiration reached `failed`; upload intent replay returned the same IDs.
- [x] Two temporary organizations proved API membership denial and direct `roliq_app` RLS isolation.
- [x] A controlled LocalStack outage persisted outbox attempts; recovery published the event to SQS.
- [x] Acceptance fixtures and temporary Clerk identities were removed after the run.

Run the integration evidence again with:

```powershell
.\scripts\acceptance\phase1.ps1
```

## Remaining release acceptance

- [ ] Exercise interactive Clerk sign-up and sign-in in a browser, then confirm protected-route redirects and bootstrap UI behavior.
- [ ] Perform desktop and mobile visual smoke testing for landing, authentication, onboarding, dashboard, profile, and resume screens.
- [ ] Run browser accessibility smoke testing for keyboard navigation, focus order, labels, landmarks, and obvious contrast issues.

The in-app browser host was unavailable during the 2026-06-24 closeout because required sandbox metadata was not supplied, so these checks are intentionally not marked complete. This is an external verification blocker, not a known application defect.

## Phase 2 gate

Phase 2 contract design, ADRs, Resume Document schema work, and synthetic parsing fixtures may begin. Phase 2 runtime implementation should begin after the three interactive browser checks above pass and this file is marked complete.
