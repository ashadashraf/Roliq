# Phase 1 security model

## Trust boundaries

- Clerk handles interactive authentication; the API trusts only cryptographically verified OIDC claims.
- `iss + sub` is the immutable external identity. Verified email is contact data, not an authorization key.
- Organization selection is never accepted without a database membership check.
- The application role cannot bypass RLS. Migration and worker credentials are separate and must not be exposed to web or API containers.
- S3 objects begin in quarantine and are not exposed by public ACLs or application download routes.

## Controls

- Short-lived bearer tokens are held by Clerk and are not persisted in browser storage by Roliq.
- Strict CORS, body limits, security headers, JSON shape validation, per-IP Redis rate limiting, request IDs, structured logs, and generic tenant-denial responses are enabled.
- PDF and DOCX are capped at 10 MB. File name, declared MIME, checksum metadata, magic bytes, DOCX structure, and ClamAV results must agree.
- Audit events record actor, organization, action, resource, request ID, and safe metadata only.

## Known Phase 1 limits

- Multi-region residency, user-facing data export/deletion workflows, WAF policy, KMS key policy, and formal compliance evidence are deployment work, not implied by this repository.
- AI processing and application automation are absent, so no consent is inferred for either activity.
