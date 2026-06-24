# Development Handoff Guide

Use this guide when resuming Roliq without prior conversation context.

## Start every session here

1. Read `docs/project-plan/README.md`.
2. Read `docs/project-plan/PROGRESS.md`.
3. Read the active phase file and its exit criteria.
4. Read accepted ADRs in `docs/adr` and relevant runbooks.
5. Inspect `git status` before editing and preserve changes that are not yours.
6. Select one next action, mark it `IN_PROGRESS`, and keep scope inside the active phase.

Suggested fresh-session instruction:

> Read `docs/project-plan/README.md`, `PROGRESS.md`, the active phase plan, and accepted ADRs. Inspect the repository and current Git state. Continue the first unblocked next action, update tests and contracts, then update `PROGRESS.md` with evidence before stopping.

## Repository map

- Product UI: `apps/web`
- Product API commands: `services/api/cmd`
- Go domain/platform code: `internal`
- Background worker: `workers/background`
- Database migration/query sources: `database`
- OpenAPI, generated client, and events: `packages/contracts`, `packages/api-client`
- Local/cloud deployment structure: `compose.yaml`, `infrastructure`
- CI: `.github/workflows/ci.yml`
- Architecture/security/runbooks: `docs/architecture`, `docs/adr`, `docs/runbooks`

## Standard verification

```powershell
$env:GOCACHE="$PWD\.cache\go-build"
go test ./...
go vet ./...
pnpm typecheck
pnpm test
pnpm lint
pnpm format:check
$env:NEXT_PUBLIC_API_URL='http://localhost:8080'
$env:NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY='<development publishable key>'
pnpm build
docker compose --env-file .env.compose.example config --quiet
```

For full integration acceptance, configure real development secrets in ignored environment files, start the complete stack, and run:

```powershell
.\scripts\acceptance\phase1.ps1
```

The script creates and deletes non-personal Clerk development identities and briefly restarts LocalStack to prove queue recovery. Browser visual/accessibility checks remain a separate manual gate. Never place real keys in commands that will be committed or copied into progress files.

## Engineering invariants

- Extend OpenAPI/event/JSON schemas before implementations and regenerate clients.
- Add append-only migrations; never rewrite an applied migration.
- Pass `organization_id` explicitly and verify membership before tenant access.
- Preserve original and prior resume versions.
- Put long-running work behind queues/outbox and make consumers idempotent.
- Keep resume/job text untrusted and out of logs.
- Do not add later-phase screens backed by fake production behavior.
- Document material architectural changes as ADRs.

## Before ending a session

1. Run verification proportional to the change.
2. Update the active phase checklist.
3. Update `PROGRESS.md` with dated commands, results, blockers, and next action.
4. Update API/event schemas, migrations, `.env.example`, runbooks, and architecture docs when affected.
5. Leave at most one clearly identified `IN_PROGRESS` task.
