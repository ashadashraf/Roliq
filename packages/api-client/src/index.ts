import type { components } from "./schema";

export type UUID = string;
export type SessionContext = components["schemas"]["SessionContext"];
export type OnboardingProgress = components["schemas"]["Onboarding"];
export type Experience = components["schemas"]["Experience"];
export type Education = components["schemas"]["Education"];
export type CareerProfile = components["schemas"]["CareerProfile"];
export type ResumeRecord = components["schemas"]["Resume"];
export type DashboardSummary = components["schemas"]["Dashboard"];
export type UploadIntent = components["schemas"]["UploadIntent"];

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
    public readonly requestId?: string,
    public readonly fields?: Record<string, string>,
  ) {
    super(message);
  }
}

export class RoliqClient {
  constructor(
    private readonly baseUrl: string,
    private readonly getToken: () => Promise<string | null>,
    private readonly getOrganizationId?: () => string | undefined,
  ) {}

  private async request<T>(path: string, init: RequestInit = {}): Promise<T> {
    const token = await this.getToken();
    if (!token)
      throw new ApiError(
        401,
        "authentication_required",
        "Please sign in to continue.",
      );
    const headers = new Headers(init.headers);
    headers.set("Authorization", `Bearer ${token}`);
    headers.set("Accept", "application/json");
    if (init.body) headers.set("Content-Type", "application/json");
    const organizationId = this.getOrganizationId?.();
    if (organizationId) headers.set("X-Organization-ID", organizationId);
    const response = await fetch(`${this.baseUrl}${path}`, {
      ...init,
      headers,
    });
    if (!response.ok) {
      const problem = await response.json().catch(() => ({}));
      throw new ApiError(
        response.status,
        problem.code ?? "request_failed",
        problem.detail ?? "The request could not be completed.",
        problem.requestId,
        problem.fields,
      );
    }
    return response.status === 204 ? (undefined as T) : response.json();
  }

  bootstrap = () =>
    this.request<SessionContext>("/v1/session/bootstrap", { method: "POST" });
  me = () => this.request<SessionContext>("/v1/me");
  dashboard = () => this.request<DashboardSummary>("/v1/dashboard");
  getOnboarding = () => this.request<OnboardingProgress>("/v1/onboarding");
  updateOnboarding = (data: Partial<OnboardingProgress>) =>
    this.request<OnboardingProgress>("/v1/onboarding", {
      method: "PATCH",
      body: JSON.stringify(data),
    });
  getProfile = () => this.request<CareerProfile>("/v1/career-profile");
  saveProfile = (profile: CareerProfile) =>
    this.request<CareerProfile>("/v1/career-profile", {
      method: "PUT",
      body: JSON.stringify(profile),
    });
  listResumes = () => this.request<{ items: ResumeRecord[] }>("/v1/resumes");
  createUpload = (data: {
    fileName: string;
    contentType: string;
    sizeBytes: number;
    checksumSha256: string;
  }) =>
    this.request<UploadIntent>("/v1/resume-uploads", {
      method: "POST",
      headers: { "Idempotency-Key": crypto.randomUUID() },
      body: JSON.stringify(data),
    });
  completeUpload = (uploadId: UUID) =>
    this.request<ResumeRecord>(`/v1/resume-uploads/${uploadId}/complete`, {
      method: "POST",
    });
}
