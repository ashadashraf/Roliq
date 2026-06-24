package contracts

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const ResumeDocumentV1 = "resume-document.v1"

var eventFiles = map[string]string{
	"resume.upload.completed.v1": "events/resume.upload.completed.v1.json",
	"resume.parse.requested.v1":  "events/resume.parse.requested.v1.json",
	"resume.parse.completed.v1":  "events/resume.parse.completed.v1.json",
	"resume.parse.failed.v1":     "events/resume.parse.failed.v1.json",
}

//go:embed resume/*.json common/*.json events/*.json
var schemaFiles embed.FS

type Validator struct {
	resume *jsonschema.Schema
	events map[string]*jsonschema.Schema
}

func NewValidator() (*Validator, error) {
	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat()

	if err := fs.WalkDir(schemaFiles, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		contents, err := schemaFiles.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded schema %s: %w", path, err)
		}
		document, err := jsonschema.UnmarshalJSON(bytes.NewReader(contents))
		if err != nil {
			return fmt.Errorf("decode embedded schema %s: %w", path, err)
		}
		identifier, err := schemaIdentifier(contents)
		if err != nil {
			return fmt.Errorf("identify embedded schema %s: %w", path, err)
		}
		if err := compiler.AddResource(identifier, document); err != nil {
			return fmt.Errorf("register embedded schema %s: %w", path, err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	resume, err := compiler.Compile("https://contracts.roliq.com/resume/resume-document.v1.schema.json")
	if err != nil {
		return nil, fmt.Errorf("compile resume document schema: %w", err)
	}
	events := make(map[string]*jsonschema.Schema, len(eventFiles))
	for eventType, path := range eventFiles {
		contents, err := schemaFiles.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read event schema %s: %w", eventType, err)
		}
		identifier, err := schemaIdentifier(contents)
		if err != nil {
			return nil, fmt.Errorf("identify event schema %s: %w", eventType, err)
		}
		schema, err := compiler.Compile(identifier)
		if err != nil {
			return nil, fmt.Errorf("compile event schema %s: %w", eventType, err)
		}
		events[eventType] = schema
	}
	return &Validator{resume: resume, events: events}, nil
}

func (v *Validator) ValidateResumeDocument(data []byte) error {
	value, err := decodeJSON(data)
	if err != nil {
		return fmt.Errorf("decode resume document: %w", err)
	}
	if err := v.resume.Validate(value); err != nil {
		return fmt.Errorf("validate %s: %w", ResumeDocumentV1, err)
	}
	return nil
}

func (v *Validator) ValidateEvent(eventType string, data []byte) error {
	schema, ok := v.events[eventType]
	if !ok {
		return fmt.Errorf("unsupported event contract %q", eventType)
	}
	value, err := decodeJSON(data)
	if err != nil {
		return fmt.Errorf("decode %s: %w", eventType, err)
	}
	if err := schema.Validate(value); err != nil {
		return fmt.Errorf("validate %s: %w", eventType, err)
	}
	return nil
}

func schemaIdentifier(data []byte) (string, error) {
	var header struct {
		ID string `json:"$id"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return "", err
	}
	if strings.TrimSpace(header.ID) == "" {
		return "", fmt.Errorf("$id is required")
	}
	return header.ID, nil
}

func decodeJSON(data []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, fmt.Errorf("multiple JSON values are not allowed")
		}
		return nil, err
	}
	return value, nil
}
