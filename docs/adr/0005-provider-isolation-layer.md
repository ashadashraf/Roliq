# ADR 0005: Provider isolation layer

**Status:** Accepted

## Context

LLM and OCR providers expose different request formats, response schemas, confidence models, error taxonomies, streaming behavior, and retention controls. Allowing those types into product contracts would couple the editor, orchestration, and database to one provider and make safe replacement difficult.

## Decision

All AI and OCR integrations implement Roliq-owned interfaces beneath `services/ai/internal/providers`. The provider-neutral Resume Document contract remains the only structured parsing output accepted by product code.

The foundation directory is:

```text
services/ai/internal/providers/
|-- interfaces/
|-- fake/
|-- openai/
|-- anthropic/
`-- gemini/
```

OCR providers follow the same interface boundary. Provider directories may exist as documented extension points, but no real provider SDK, credentials, HTTP calls, or provider-specific runtime logic is added during foundation hardening.

The interfaces use Roliq-owned request/result/error types:

- parser input identifies a checksummed private source or extraction artifact plus schema/prompt versions and budgets;
- parser output is a validated `resume-document.v1` value plus provider-neutral usage and trace metadata;
- OCR output is ordered page text with provenance and provider-neutral confidence;
- errors are classified as validation, authentication, rate limit, timeout, unavailable, budget exceeded, content rejected, or permanent input failure.

Provider-native objects are converted inside the adapter and discarded. They cannot appear in queue contracts, OpenAPI, database JSON, web types, audit metadata, or logs. A mapping is successful only after validation against the canonical schema.

The deterministic fake parser and fake OCR adapters are mandatory first implementations. They read only synthetic fixture data, make no network calls, and enable contract/integration tests before a provider is selected.

## Enforcement

- The AI package has no provider SDK dependencies until the corresponding provider ADR is accepted.
- Imports from provider adapter packages are prohibited outside the provider composition boundary.
- Cross-language tests validate the same canonical schema; adapters do not maintain schema copies.
- Provider response fixtures, when later added, are sanitized and remain adapter-local.

## Consequences

- New providers require mapping and contract tests but do not change product APIs or persistence.
- The shared contract may expose fewer provider-specific features; deliberate contract evolution is preferred to accidental coupling.
- Fake adapters can prove orchestration semantics, but they do not establish real-provider quality; ADR 0004 evaluation remains required.

## Rejected alternatives

- Storing raw provider responses as the product model creates vendor lock-in and inconsistent validation.
- A generic `dict` boundary without Roliq-owned types merely hides coupling until runtime.
- Installing every provider SDK in advance increases supply-chain and configuration risk without product value.
