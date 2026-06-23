# Resume processing runbook

## State progression

`pending → uploaded → scanning → ready`

Unsafe or structurally invalid files become `rejected`. Storage or scanner failures become `failed`; they never become ready through retries alone.

## Investigation

1. Locate the resume ID in structured worker logs.
2. Inspect `resumes.status`, `file_objects.scan_status`, and `file_objects.scan_detail` using privileged operational access.
3. Confirm ClamAV health and signature freshness.
4. Confirm S3 object metadata contains the expected SHA-256 value and that the object remains private.
5. Inspect unpublished outbox rows and the SQS dead-letter policy in the target environment.

Do not download candidate files to an unmanaged workstation. Do not manually set a resume to `ready`; require a new upload after a rejected or failed scan.
