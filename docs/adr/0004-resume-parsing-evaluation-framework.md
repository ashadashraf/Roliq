# ADR 0004: Resume parsing evaluation framework

**Status:** Accepted

## Context

Schema-valid output can still omit facts, misclassify dates, or invent unsupported details. Provider, model, prompt, extraction, and normalization changes therefore need a repeatable quality gate before they can affect users. Production resumes cannot be committed to the repository or become an informal test corpus.

## Decision

Roliq will maintain a version-controlled golden dataset under `fixtures/resumes`. The initial dataset is entirely synthetic, original to this repository, and contains no real personal information or copied resumes.

Each scenario directory contains:

- `resume.pdf`: a deterministic, render-verified synthetic source document;
- `expected_resume_document_v1.json`: the canonical expected structured result;
- `evaluation_notes.md`: scenario intent, layout characteristics, ambiguity, and expected review points;
- `evaluation_metadata.json`: machine-readable tags, page count, source checksum, and metric weights;
- `fake_ocr_output.json`: deterministic offline page text for fake-adapter tests.

The initial scenarios are senior backend engineer, frontend engineer, data scientist, fresh graduate, product manager, and career switcher. New extraction or parsing behavior must add a fixture when it introduces an unrepresented layout or semantic edge case.

## Metrics

Evaluation compares a candidate `resume-document.v1` output with the expected document after deterministic normalization.

| Metric                            | Definition                                                                                                                                |
| --------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| Experience extraction accuracy    | Macro F1 over experience entities matched by normalized company, role, and overlapping dates, plus field-level F1 inside matched entities |
| Skill extraction accuracy         | Set precision, recall, and F1 over normalized skill names; aliases are allowed only through a versioned normalization table               |
| Education extraction accuracy     | Macro F1 over institutions/degrees/fields and field-level accuracy for matched entities                                                   |
| Certification extraction accuracy | Set/entity F1 over certification name and issuer                                                                                          |
| Date extraction accuracy          | Exact normalized-value accuracy, with separately reported partial-date compatibility                                                      |
| Hallucination rate                | Unsupported predicted factual fields divided by all predicted factual fields; support requires matching source evidence                   |
| Evidence coverage                 | Predicted factual fields with valid, source-matching evidence divided by all predicted factual fields                                     |
| Review-required rate              | Factual fields marked low-confidence, warned, unsupported, or `unreviewed` divided by all predicted factual fields                        |

Entity matching and normalization rules are versioned with the evaluator. Aggregate results are reported both macro-averaged across scenarios and per fixture so a common layout cannot hide a regression in a difficult one.

## Regression policy

- Every prompt, model, provider adapter, extraction, OCR, normalization, and schema-mapping change runs the full golden suite.
- A candidate cannot increase hallucination count, reduce evidence coverage, or reduce any core macro F1 by more than two percentage points without an approved baseline update.
- Any regression in a safety-critical fixture fails regardless of aggregate score.
- Baseline changes require reviewed expected-output diffs and an explanation in the fixture notes; a new model score alone cannot rewrite the golden truth.
- Provider evaluations record model identifier, adapter version, prompt version, schema version, evaluator version, latency, token usage, estimated cost, and timestamp.

The fake provider must achieve exact fixture identity. Real providers are evaluated later and are never called by the default offline test suite.

## Future production-derived evaluation

Production-derived examples may be added only through a separately approved consent, de-identification, access-control, retention, and deletion process. They must not be committed to the public development corpus. Until that process exists, repository evaluation remains synthetic.

## Consequences

- Quality becomes a measured release property rather than a subjective prompt review.
- The repository carries more test artifacts, but provider changes become comparable and reversible.
- The golden dataset tests known cases; separate adversarial, privacy, load, and live canary testing remains necessary.
