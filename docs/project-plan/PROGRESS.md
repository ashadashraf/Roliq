# Development Progress

This file is the authoritative live ledger. Update it in the same change as implementation work.

Last updated: 2026-06-24

## Current milestone

- Primary milestone: Phase 2 planning
- Parallel milestone: Phase 1 interactive browser closeout
- Status: IN_PROGRESS - planning only; Phase 2 runtime has not started
- Phase 2 current task: finish Python/fixture/PDF verification for P2-010 through P2-013 and P2-015
- Phase 1 blocker: browser-host availability for desktop/mobile/authentication/accessibility smoke tests

## Phase status

| Phase | Status      | Completion note                                                                  |
| ----- | ----------- | -------------------------------------------------------------------------------- |
| 1     | IN_PROGRESS | Implementation and automated acceptance complete; interactive browser QA remains |
| 2     | IN_PROGRESS | Planning active; ADRs/backlog created; runtime not started                       |
| 3     | NOT_STARTED | Scope and exit criteria recorded                                                 |
| 4     | NOT_STARTED | Scope and exit criteria recorded                                                 |
| 5     | NOT_STARTED | Scope and exit criteria recorded                                                 |
| 6     | NOT_STARTED | Scope and exit criteria recorded                                                 |

## Latest verification

Run on 2026-06-24:

| Check                                     | Result  | Notes                                                                         |
| ----------------------------------------- | ------- | ----------------------------------------------------------------------------- |
| Go test, vet, and formatting              | PASS    | Includes Go contract validation across all golden documents and events        |
| pnpm lint, typecheck, test, build         | PASS    | Includes six TypeScript contract tests and production Next.js build           |
| Root `pnpm format:check`                  | BLOCKED | Relevant files are formatted; inaccessible temporary Python cache blocks scan |
| Python source compilation                 | PASS    | AI foundation, fake adapters, tests, and fixture tools compile                |
| Python pytest and fixture validator       | BLOCKED | Requires clean dependency install after temporary cache ACL cleanup           |
| Synthetic PDF visual rendering            | BLOCKED | Requires the same clean local PDF-tool installation                           |
| `docker compose ... config --quiet`       | PASS    | Compose model parses                                                          |
| Fresh-volume full Compose startup         | PASS    | Disposable project; all services healthy/running; volumes removed afterward   |
| API/web health and three container images | PASS    | API ready 200, web 200, API/worker/web images built                           |
| Migration and database roles              | PASS    | Goose v1 applied; app role cannot bypass RLS                                  |
| Linux-container `sqlc compile`            | PASS    | Same image/version as CI                                                      |
| Real Clerk/OIDC integration               | PASS    | JWT template, audience, verified email, discovery/JWKS, bootstrap             |
| `scripts/acceptance/phase1.ps1`           | PASS    | Both onboarding paths, resume security, RLS, outbox retry/delivery            |
| Interactive browser/accessibility         | BLOCKED | In-app browser host unavailable because required sandbox metadata was missing |

## Next actions

1. [ ] Remove the disposable `.cache/python` directory with explicit approval, install `services/ai[dev]` cleanly, and run Python/fixture/PDF checks.
2. [ ] Mark P2-010 through P2-013 and P2-015 complete only after those checks pass.
3. [ ] Assign product/security/platform owners and target dates to P2-D01 through P2-D04.
4. [ ] Restore the in-app browser host and finish the three Phase 1 interactive checks.
5. [ ] Do not start migrations, FastAPI/runtime parsing, OCR, or real-provider integration until their recorded gates are satisfied.

## Known workspace state

- Git branch: `main`.
- Latest observed commit before this closeout: `7a17823 fix: stabilize Docker startup and resume uploads`.
- Existing user changes observed before planning/closeout work: staged `.dockerignore` and modified `apps/web/next-env.d.ts`; preserve them.
- `.env`, Clerk generated state, build output, real resumes, and provider credentials must not be committed.
- The configured Clerk development tenant now contains the required `roliq-api` JWT template; its non-secret claims are documented in the root README.
- Environment: Windows/PowerShell. Git reads may require `-c safe.directory=C:/Users/user/projects/WorkSpace/JobAutomationSaas`.

## Update log

### 2026-06-24 - Phase 2 provider-free foundation hardening

- Accepted ADR 0004 for measurable golden-dataset evaluation and regression prevention.
- Accepted ADR 0005 for provider-native type isolation behind Roliq-owned interfaces.
- Accepted ADR 0006 for hard document, retry, token, monetary, cancellation, duplicate, and queue safeguards.
- Added the canonical `resume-document.v1` Draft 2020-12 schema and pointer-only requested/completed/failed parsing events.
- Added Go, Python, and TypeScript validation packages that load the same schema files without copies.
- Added six original synthetic PDF golden scenarios with expected documents, evaluation notes/metadata, and deterministic fake OCR output.
- Added offline fake parser/OCR interfaces and adapters; no provider SDKs, external calls, runtime parser, or FastAPI service were added.
- Go tests/vet, TypeScript contract tests, all pnpm gates, production build, Python syntax compilation, and diff checks pass.
- Python pytest, evidence-aware fixture validation, PDF render inspection, and the root format command remain pending because the temporary local dependency cache has restrictive ACLs.

### 2026-06-24 - Phase 2 planning started

- Marked Phase 2 planning `IN_PROGRESS` while keeping runtime implementation explicitly not started.
- Accepted ADR 0002 for the evidence-backed Resume Document and immutable version model.
- Accepted ADR 0003 for asynchronous parsing, pointer-only events, and no Python product-database access.
- Added `PHASE_02_DECISIONS.md` with four open vendor, OCR, identity, consent, and retention gates.
- Added `PHASE_02_BACKLOG.md` with stable task IDs, dependencies, milestone exits, and acceptance evidence.
- Selected P2-010, P2-011, and P2-013 as the first provider-independent implementation slice after planning approval.

### 2026-06-24 - Phase 1 automated acceptance

- Restored the formatting gate by excluding generated Clerk state and formatting Compose.
- Added LF enforcement for Linux shell/config files and fixed LocalStack startup on Windows checkouts.
- Started the full stack and proved a separate disposable clean-volume startup, migrations, health checks, roles, and all container builds.
- Configured the documented Clerk development JWT template and proved real token issuance plus provider-neutral validation.
- Added `scripts/acceptance/phase1.ps1` for repeatable auth, bootstrap, onboarding, persisted dashboard, resume security, RLS, and outbox checks.
- Fixed ClamAV protocol response normalization so the terminal NUL byte cannot break PostgreSQL rejection persistence; added a regression test.
- Passed Go, pnpm, OpenAPI, Compose, Docker, and Linux `sqlc` gates.
- Browser smoke testing remains blocked by the unavailable browser host and is the only Phase 1 closeout item.

### 2026-06-23 - Durable planning

- Audited Phase 1 implementation and reran quality gates.
- Classified Phase 1 as implemented but not acceptance-closed.
- Added the roadmap, closeout checklist, Phase 2 plan, progress ledger, and handoff guide.

## Entry template

```markdown
### YYYY-MM-DD - short task name

- Status: IN_PROGRESS | BLOCKED | COMPLETE
- Objective:
- Changed:
- Contracts/migrations:
- Verification commands and results:
- Remaining risks/blockers:
- Next exact action:
```
