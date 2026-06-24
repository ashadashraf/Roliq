import Ajv2020, {
  type ErrorObject,
  type ValidateFunction,
} from "ajv/dist/2020.js";
import addFormats from "ajv-formats";

import artifactPointerV1 from "../common/artifact-pointer.v1.schema.json";
import resumeParseCompletedV1 from "../events/resume.parse.completed.v1.json";
import resumeParseFailedV1 from "../events/resume.parse.failed.v1.json";
import resumeParseRequestedV1 from "../events/resume.parse.requested.v1.json";
import resumeUploadCompletedV1 from "../events/resume.upload.completed.v1.json";
import resumeDocumentV1 from "../resume/resume-document.v1.schema.json";

export const RESUME_DOCUMENT_V1 = "resume-document.v1" as const;

export type ValidationResult =
  | { valid: true; errors: [] }
  | { valid: false; errors: ErrorObject[] };

const ajv = new Ajv2020({ allErrors: true, strict: true });
addFormats(ajv);
ajv.addSchema(artifactPointerV1);

const resumeValidator = ajv.compile(resumeDocumentV1);
const eventValidators: Readonly<Record<string, ValidateFunction>> = {
  "resume.upload.completed.v1": ajv.compile(resumeUploadCompletedV1),
  "resume.parse.requested.v1": ajv.compile(resumeParseRequestedV1),
  "resume.parse.completed.v1": ajv.compile(resumeParseCompletedV1),
  "resume.parse.failed.v1": ajv.compile(resumeParseFailedV1),
};

function result(validator: ValidateFunction, value: unknown): ValidationResult {
  if (validator(value)) {
    return { valid: true, errors: [] };
  }
  return { valid: false, errors: [...(validator.errors ?? [])] };
}

export function validateResumeDocument(value: unknown): ValidationResult {
  return result(resumeValidator, value);
}

export function validateEvent(
  eventType: string,
  value: unknown,
): ValidationResult {
  const validator = eventValidators[eventType];
  if (!validator) {
    throw new Error(`Unsupported event contract: ${eventType}`);
  }
  return result(validator, value);
}

export function assertResumeDocument(value: unknown): void {
  const validation = validateResumeDocument(value);
  if (!validation.valid) {
    throw new Error(ajv.errorsText(validation.errors, { separator: "; " }));
  }
}
