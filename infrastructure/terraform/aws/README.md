# AWS reference infrastructure

Phase 1 does not provision cloud resources. This directory reserves the Phase 6 Terraform root and documents its required modules: networking, EKS, RDS PostgreSQL, ElastiCache, private S3/CloudFront, ALB, Route53, ECR, SNS/SQS with dead-letter queues, SES, Secrets Manager, CloudWatch/OTLP, and OpenSearch.

Application containers depend only on Kubernetes, PostgreSQL, Redis, S3-compatible object storage, OIDC/JWKS, SQS-compatible publishing, and OTLP. Provider resources must be translated into these interfaces instead of imported into domain packages.
