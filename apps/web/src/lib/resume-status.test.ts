import { describe, expect, it } from "vitest";
import { resumeStatusLabel } from "./resume-status";

describe("resumeStatusLabel", () => {
  it("uses explicit security language for pre-ready states", () => {
    expect(resumeStatusLabel("uploaded")).toBe("Queued for security scan");
    expect(resumeStatusLabel("ready")).toBe("Securely stored");
  });
});
