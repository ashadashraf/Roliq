import type { ResumeRecord } from "@roliq/api-client";

const labels: Record<ResumeRecord["status"], string> = {
  pending: "Waiting for upload",
  uploaded: "Queued for security scan",
  scanning: "Security scan in progress",
  ready: "Securely stored",
  rejected: "Rejected for safety",
  failed: "Processing failed",
};

export function resumeStatusLabel(status: ResumeRecord["status"]): string {
  return labels[status];
}
