# ADR 0001: Modular product API with isolated high-risk runtimes

**Status:** Accepted

Roliq begins with one Go product API so tenant provisioning, onboarding, profiles, audit logs, and outbox events remain transactional. Python AI workloads and Playwright automation are separate services because they have distinct dependencies, scaling profiles, and security risk. Domain boundaries remain internal modules until operational evidence justifies extraction.
