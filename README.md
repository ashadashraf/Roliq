# Roliq

Roliq is a tenant-safe career workspace and the production foundation for a job application platform. Phase 1 deliberately stops at authenticated onboarding, structured career profiles, secure resume intake, and auditable background processing. It does **not** parse resumes, recommend jobs, generate documents, charge users, or submit applications.

Development roadmap, phase plans, live progress, and context-recovery instructions are maintained in [`docs/project-plan`](docs/project-plan/README.md).

## Architecture

- `apps/web` — Next.js App Router, Clerk, TanStack Query, Tailwind CSS
- `services/api` — Go/Echo modular API with provider-neutral OIDC verification
- `services/ai` — provider-neutral Phase 2 contracts and offline fake adapters; no runtime or provider SDKs yet
- `workers/background` — isolated scanning and transactional-outbox worker
- `database` — PostgreSQL migrations and sqlc query sources
- `packages/contracts` — OpenAPI and versioned event schemas
- `infrastructure` — local Docker, Kubernetes base, and AWS reference material

The API is a modular monolith because Phase 1 transactions cross identity, tenant provisioning, onboarding, profile, audit, and outbox boundaries. AI and browser automation remain separate runtime and security boundaries.

## Prerequisites

- Node.js 22+, pnpm 10+
- Go 1.25+
- Docker Desktop with Compose v2
- A Clerk development application

## Clerk and OIDC setup

1. Enable email verification and optionally Google sign-in in Clerk.
2. Create a JWT template named `roliq-api`.
3. Use the following claims. Roliq keys identities by `iss + sub`, never by email.

```json
{
  "aud": "roliq-api",
  "email": "{{user.primary_email_address}}",
  "email_verified": "{{user.email_verified}}",
  "name": "{{user.full_name}}"
}
```

4. Copy `apps/web/.env.example` to `apps/web/.env.local` and add the Clerk keys.
5. Copy `.env.compose.example` to `.env`. Set `OIDC_ISSUER_URL` to the token `iss` claim and `OIDC_JWKS_URL` to the provider JWKS endpoint when discovery is unavailable.

The Go API imports no Clerk SDK. A different OIDC provider works by changing issuer, JWKS, audience, and client configuration.

## Local development

```powershell
pnpm install
docker compose up postgres redis localstack clamav migrate
$env:DATABASE_URL='postgres://roliq_app:roliq_app@localhost:55432/roliq?sslmode=disable'
$env:REDIS_URL='redis://localhost:6379/0'
# Add the OIDC and S3 values from services/api/.env.example.
go run ./services/api/cmd/api
```

In another terminal:

```powershell
go run ./workers/background/cmd/worker
pnpm --filter @roliq/web dev
```

Alternatively, configure `.env` and run the complete stack with `docker compose up --build`.
The optional local telemetry stack runs with `docker compose --profile observability up`; set `OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318` in `.env` to export traces.

## Quality gates

```powershell
gofmt -w internal services workers
go test ./...
go vet ./...
pnpm lint
pnpm typecheck
pnpm test
pnpm build
```

With the full local stack running and a development Clerk tenant configured, run the disposable Phase 1 integration suite:

```powershell
.\scripts\acceptance\phase1.ps1
```

The suite creates and deletes non-personal Clerk test identities, exercises both onboarding paths, secure resume processing, RLS, and outbox recovery, and removes its database and object-storage fixtures.

Database changes are append-only migrations. Run them with `MIGRATION_DATABASE_URL` and `go run ./services/api/cmd/migrate`. Never run the API using the migration-owner credentials.

## Security model

- Business records are scoped by `organization_id`, composite foreign keys, explicit query predicates, and PostgreSQL row-level security.
- New users receive one personal organization and owner membership in an idempotent serializable transaction.
- Resume bytes upload directly to a quarantine prefix using a 15-minute presigned URL. Size, content type, SHA-256 metadata, file signature, DOCX structure, and malware status are verified before a resume becomes `ready`.
- Application events and audit records commit in the same transaction as their domain change. Resume contents and access tokens are never written to audit metadata.

See `docs/architecture/system.md`, `docs/architecture/security.md`, and `docs/runbooks/resume-processing.md` for operational detail.
