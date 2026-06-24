package background

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestValidateDocumentPDF(t *testing.T) {
	if err := validateDocument([]byte("%PDF-1.7\nbody"), "application/pdf"); err != nil {
		t.Fatal(err)
	}
	if err := validateDocument([]byte("not a pdf"), "application/pdf"); err == nil {
		t.Fatal("expected invalid signature")
	}
}
func TestValidateDocumentDOCX(t *testing.T) {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	file, err := writer.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = file.Write([]byte("<w:document/>"))
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := validateDocument(buffer.Bytes(), "application/vnd.openxmlformats-officedocument.wordprocessingml.document"); err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeClamAVDetailRemovesProtocolTerminator(t *testing.T) {
	detail := normalizeClamAVDetail([]byte("stream: Eicar-Signature FOUND\x00\n"))
	if detail != "stream: Eicar-Signature FOUND" {
		t.Fatalf("unexpected detail %q", detail)
	}
	if bytes.ContainsRune([]byte(detail), '\x00') {
		t.Fatal("ClamAV detail must be safe to persist as PostgreSQL text")
	}
}
