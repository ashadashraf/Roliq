# Roliq Project Plan

This folder is the durable source of truth for product scope, implementation phases, current progress, and development handoff. It is designed so a new engineer or coding agent can continue without access to previous conversations.

## Current state

- Active milestones: **Phase 1 interactive browser closeout** and **Phase 2 planning**
- Phase 1 implementation and automated acceptance: **complete in scope**
- Phase 1 release acceptance: **interactive browser QA remains**
- Phase 2: **planning active; runtime not started**
- Recommendation: complete Clerk UI, responsive, and accessibility browser smoke tests before Phase 2 runtime work. Phase 2 contract, ADR, and schema design may proceed in parallel.

## Documents

- [MASTER_PLAN.md](MASTER_PLAN.md) — product direction, architecture boundaries, roadmap, and global completion rules.
- [PHASES.md](PHASES.md) — deliverables and exit criteria for Phases 1 through 6.
- [PHASE_01_CLOSEOUT.md](PHASE_01_CLOSEOUT.md) — implemented scope, current evidence, and the remaining acceptance checklist.
- [PHASE_02_PLAN.md](PHASE_02_PLAN.md) — implementation-ready plan for parsing, structured resume data, editing, preferences, and versioning.
- [PHASE_02_DECISIONS.md](PHASE_02_DECISIONS.md) — accepted architecture and unresolved vendor, identity, consent, and retention gates.
- [PHASE_02_BACKLOG.md](PHASE_02_BACKLOG.md) — stable task IDs, dependencies, milestone exits, and acceptance evidence.
- [PHASE_02_FOUNDATION_HARDENING.md](PHASE_02_FOUNDATION_HARDENING.md) — provider-free foundation scope, implemented artifacts, and final verification checklist.
- [PROGRESS.md](PROGRESS.md) — the authoritative live status ledger and next actions.
- [HANDOFF.md](HANDOFF.md) — exact instructions for resuming development in a new session.

Architecture decisions remain in [`../adr`](../adr). Operational documentation remains in [`../runbooks`](../runbooks).

## Update rules

1. Read this file, `PROGRESS.md`, the active phase plan, and accepted ADRs before changing code.
2. Set one task to `IN_PROGRESS` in `PROGRESS.md` before implementation.
3. Update evidence and remaining work after each meaningful change.
4. Mark a phase `COMPLETE` only when every exit criterion has evidence.
5. Record design changes as ADRs; do not silently rewrite accepted architecture.
6. Keep secrets, personal resumes, and production data out of planning files.

Status values are `NOT_STARTED`, `IN_PROGRESS`, `BLOCKED`, and `COMPLETE`.
