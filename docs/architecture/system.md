# System architecture

## Runtime boundaries

The web app obtains a short-lived Clerk JWT using the `roliq-api` template and sends it as a bearer token. The API validates it through OIDC discovery or a configured JWKS URL, then resolves application identity and organization membership from PostgreSQL.

The API owns synchronous product transactions. Each mutation scopes its database transaction with `app.user_id` and `app.organization_id`; PostgreSQL RLS provides defense in depth. A dedicated worker role has `BYPASSRLS` only because it must claim outbox and scan jobs across organizations. Worker queries always carry organization identifiers into subsequent updates.

Resume uploads travel directly from the browser to S3. The API stores declared metadata before issuing a presigned URL, confirms stored metadata with `HeadObject`, and only then emits `resume.upload.completed.v1`. The worker validates and scans the file before emitting the completed or rejected event.

## AWS reference deployment

| Capability        | AWS reference           | Portable boundary                       |
| ----------------- | ----------------------- | --------------------------------------- |
| Containers        | EKS and ECR             | Kubernetes Deployments and OCI images   |
| Database          | RDS PostgreSQL          | PostgreSQL and SQL migrations           |
| Cache/rate limits | ElastiCache Redis       | Redis protocol                          |
| Files             | S3 and CloudFront       | Object store interface                  |
| Events            | SNS and SQS             | Queue publisher interface plus outbox   |
| Edge              | Route53 and ALB         | Kubernetes Ingress                      |
| Secrets           | Secrets Manager         | Environment/secret injection            |
| Telemetry         | CloudWatch through OTLP | OpenTelemetry                           |
| Search, later     | OpenSearch              | Search adapter, not included in Phase 1 |

Infrastructure provisioning is intentionally completed in Phase 6. Phase 1 provides container, configuration, health, and Kubernetes contracts without claiming that an AWS account has been provisioned.
