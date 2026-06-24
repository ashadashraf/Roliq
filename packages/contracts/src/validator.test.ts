import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

import { validateEvent, validateResumeDocument } from "./validator";

const root = resolve(import.meta.dirname, "../../..");
const json = (path: string): unknown =>
  JSON.parse(readFileSync(resolve(root, path), "utf8"));

describe("Resume Document v1", () => {
  it("validates every synthetic golden document", () => {
    const fixtures = [
      "senior_backend_engineer",
      "frontend_engineer",
      "data_scientist",
      "fresh_graduate",
      "product_manager",
      "career_switcher",
    ];
    for (const fixture of fixtures) {
      const result = validateResumeDocument(
        json(`fixtures/resumes/${fixture}/expected_resume_document_v1.json`),
      );
      expect(result, fixture).toEqual({ valid: true, errors: [] });
    }
  });

  it("rejects provider-specific unknown fields", () => {
    const document = json(
      "fixtures/resumes/fresh_graduate/expected_resume_document_v1.json",
    ) as Record<string, unknown>;
    document.providerResponse = { unsafe: true };
    expect(validateResumeDocument(document).valid).toBe(false);
  });
});

describe("pointer-only parsing events", () => {
  it.each([
    ["resume.parse.requested.v1", "resume.parse.requested.v1.valid.json"],
    ["resume.parse.completed.v1", "resume.parse.completed.v1.valid.json"],
    ["resume.parse.failed.v1", "resume.parse.failed.v1.valid.json"],
  ])("validates %s", (eventType, fileName) => {
    expect(
      validateEvent(
        eventType,
        json(`packages/contracts/testdata/events/${fileName}`),
      ),
    ).toEqual({ valid: true, errors: [] });
  });

  it("rejects a parsing event containing resume text", () => {
    const event = json(
      "packages/contracts/testdata/events/resume.parse.requested.v1.invalid-pii.json",
    );
    expect(validateEvent("resume.parse.requested.v1", event).valid).toBe(false);
  });
});
