# Phase 1 acceptance

`phase1.ps1` is the repeatable integration acceptance suite for the production foundation. It uses the real local API, PostgreSQL roles and RLS, LocalStack S3/SQS, ClamAV, and a configured Clerk development tenant.

## Prerequisites

- The complete Compose stack is running and healthy.
- `.env` contains development Clerk keys and the correct issuer/audience.
- The Clerk tenant has a JWT template named `roliq-api` with the claims documented in the root `README.md`.
- Docker Desktop can be accessed by the current process.

## Run

```powershell
.\scripts\acceptance\phase1.ps1
```

The suite creates two non-personal `+clerk_test` identities and deletes them in `finally` cleanup. It also removes their personal organizations, database records, and S3 prefixes. LocalStack is deliberately stopped once to prove outbox retry and recovery, then restarted before cleanup.

The suite verifies:

- Clerk JWT-template issuance and provider-neutral OIDC validation;
- idempotent session bootstrap and personal organization creation;
- manual and resume onboarding persistence;
- persisted career profile and dashboard data;
- API membership denial and direct PostgreSQL RLS isolation;
- PDF/DOCX acceptance, invalid signatures, checksum mismatch, EICAR detection, expiration, and upload-intent idempotency;
- transactional-outbox failure persistence, recovery, and SQS delivery.

Do not point this suite at a production Clerk tenant or production infrastructure.
