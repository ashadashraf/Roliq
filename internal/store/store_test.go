package store

import (
	"strings"
	"testing"

	"github.com/roliq/roliq/internal/model"
)

func TestAdvisoryLockKeyIsDatabaseSafeAndInputSpecific(t *testing.T) {
	first := advisoryLockKey("https://issuer.example", "user_123")
	second := advisoryLockKey("https://issuer.example", "user_456")

	if strings.ContainsRune(first, '\x00') {
		t.Fatal("advisory lock key must not contain a NUL byte")
	}
	if len(first) != 64 {
		t.Fatalf("expected a SHA-256 hex key, got %q", first)
	}
	if first == second {
		t.Fatal("different inputs must produce different lock keys")
	}
}

func TestProfileCompletion(t *testing.T) {
	years := 5.0
	profile := model.CareerProfile{Headline: "Platform engineer", Summary: "Builds reliable systems.", CountryCode: "IN", TimeZone: "Asia/Kolkata", YearsExperience: &years, Skills: []string{"Go", "PostgreSQL", "AWS"}, Experiences: []model.Experience{{Company: "Example", Title: "Engineer", StartDate: "2020-01-01"}}, Education: []model.Education{{Institution: "University"}}}
	if got := ProfileCompletion(profile); got != 100 {
		t.Fatalf("expected complete profile score 100, got %d", got)
	}
}
func TestPersonalSlugDoesNotExposeEmail(t *testing.T) {
	slug := personalSlug("Ada Lovelace", "018f1fc2-8df3-7e82-b58d-3f4fa4d22573")
	if slug != "ada-lovelace-018f1fc28d" {
		t.Fatalf("unexpected slug %q", slug)
	}
}
