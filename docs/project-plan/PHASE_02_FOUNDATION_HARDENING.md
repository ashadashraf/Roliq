# Phase 2 Foundation Hardening

Status: **IN_REVIEW - IMPLEMENTED, FINAL PYTHON/PDF VERIFICATION PENDING**

Last updated: 2026-06-24

## Scope guardrails

- No runtime parsing, FastAPI service, queue consumer, persistence migration, or orchestration was added.
- No OCR or LLM provider was selected or called.
- No OpenAI, Anthropic, Gemini, Textract, or other provider SDK was installed.
- ADR 0002 immutable versions and ADR 0003 pointer-only asynchronous events remain unchanged.

## Implemented

- [x] ADR 0004: golden dataset, metrics, evaluation, and regression policy.
- [x] ADR 0005: provider-neutral interfaces and adapter isolation.
- [x] ADR 0006: document, retry, token, monetary, cancellation, duplicate, and queue limits.
- [x] Canonical JSON Schema draft 2020-12 `resume-document.v1` with strict fields, evidence, confidence, warnings, and review state.
- [x] Shared artifact-pointer schema and versioned requested/completed/failed parsing events.
- [x] Go, Python, and TypeScript validators load the same canonical schema files.
- [x] Six original synthetic PDF scenarios with expected documents, evaluation notes/metadata, and fake OCR output.
- [x] Fixture generation, evidence-aware validation, and PDF rendering/contact-sheet tools.
- [x] Roliq-owned parser/OCR interfaces and deterministic offline fake implementations.
- [x] CI installation and tests for the Python foundation plus existing Go/TypeScript gates.

## Verification evidence

- [x] `go test ./...`
- [x] `go vet ./...`
- [x] Go contract tests validate all six golden documents and parsing events.
- [x] `pnpm lint`, `pnpm typecheck`, `pnpm test`, and `pnpm build`.
- [x] TypeScript contract suite: six tests pass.
- [x] Python source compiles with `compileall`.
- [x] Repository diff whitespace check passes.
- [ ] Python pytest suite executes successfully in a clean environment.
- [ ] Evidence-aware fixture validation tool executes successfully.
- [ ] Latest fixture PDFs are rendered and visually inspected for clipping, overlap, and legibility.
- [ ] Root `pnpm format:check` is rerun after the inaccessible local dependency cache is removed.

The four pending items are environment verification tasks. The temporary local Python dependency cache was created with ACLs that the current restricted process cannot traverse, and removal/reinstallation requires explicit approval. CI uses a clean `pip install -e "./services/ai[dev]"` path and does not rely on that cache.

## Gate before any real provider

All checks above must pass. P2-D01 through P2-D04 must also be resolved, and the offline orchestration workflow must prove duplicate protection, cancellation, budgets, and one-version creation before any provider SDK or external call is allowed.
