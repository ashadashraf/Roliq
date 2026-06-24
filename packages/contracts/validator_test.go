package contracts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestResumeFixturesValidateAcrossCanonicalSchema(t *testing.T) {
	validator, err := NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	paths, err := filepath.Glob(filepath.Join("..", "..", "fixtures", "resumes", "*", "expected_resume_document_v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 6 {
		t.Fatalf("expected 6 golden resume documents, got %d", len(paths))
	}
	for _, path := range paths {
		path := path
		t.Run(filepath.Base(filepath.Dir(path)), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := validator.ValidateResumeDocument(data); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestResumeDocumentRejectsUnknownFields(t *testing.T) {
	validator, err := NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("..", "..", "fixtures", "resumes", "fresh_graduate", "expected_resume_document_v1.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	document["providerResponse"] = map[string]any{"unsafe": true}
	invalid, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := validator.ValidateResumeDocument(invalid); err == nil {
		t.Fatal("expected unknown provider field to be rejected")
	}
}

func TestParsingEventsArePointerOnly(t *testing.T) {
	validator, err := NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"resume.parse.requested.v1": "resume.parse.requested.v1.valid.json",
		"resume.parse.completed.v1": "resume.parse.completed.v1.valid.json",
		"resume.parse.failed.v1":    "resume.parse.failed.v1.valid.json",
	}
	for eventType, fileName := range cases {
		data, err := os.ReadFile(filepath.Join("testdata", "events", fileName))
		if err != nil {
			t.Fatal(err)
		}
		if err := validator.ValidateEvent(eventType, data); err != nil {
			t.Fatalf("%s: %v", eventType, err)
		}
	}
	invalid, err := os.ReadFile(filepath.Join("testdata", "events", "resume.parse.requested.v1.invalid-pii.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := validator.ValidateEvent("resume.parse.requested.v1", invalid); err == nil {
		t.Fatal("expected event containing resume text to be rejected")
	}
}
