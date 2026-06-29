package utils

import (
	"strings"
	"testing"
)

func TestValidateUploadOriginalFileNameRejectsPathTraversal(t *testing.T) {
	cases := []string{"../../../etc/passwd", "..\\secret.pdf", "/etc/passwd", `a\b.pdf`}
	for _, name := range cases {
		if err := ValidateUploadOriginalFileName(name); !IsValidationError(err) {
			t.Fatalf("expected validation error for %q, got %v", name, err)
		}
	}
}

func TestValidateUploadOriginalFileNameAcceptsSimpleNames(t *testing.T) {
	cases := []string{"informe final.pdf", "INFORME.PDF", "report.docx"}
	for _, name := range cases {
		if err := ValidateUploadOriginalFileName(name); err != nil {
			t.Fatalf("unexpected error for %q: %v", name, err)
		}
	}
}

func TestSanitizeUploadFileName(t *testing.T) {
	got := SanitizeUploadFileName("informe final.pdf", "")
	if got != "informe-final.pdf" {
		t.Fatalf("got %q", got)
	}

	got = SanitizeUploadFileName("INFORME.PDF", "")
	if got != "INFORME.pdf" {
		t.Fatalf("got %q", got)
	}

	got = SanitizeUploadFileName("noext", ".pdf")
	if got != "noext.pdf" {
		t.Fatalf("got %q", got)
	}

	long := strings.Repeat("a", 150) + ".pdf"
	got = SanitizeUploadFileName(long, "")
	if len([]rune(strings.TrimSuffix(got, ".pdf"))) > maxUploadBaseNameLen {
		t.Fatalf("base name too long: %q", got)
	}
}

func TestBuildUniqueUploadKeyUniqueAndSafe(t *testing.T) {
	key1, stored1, orig1, err := BuildUniqueUploadKey("informe.pdf", ".pdf")
	if err != nil {
		t.Fatal(err)
	}
	key2, stored2, _, err := BuildUniqueUploadKey("informe.pdf", ".pdf")
	if err != nil {
		t.Fatal(err)
	}
	if key1 == key2 || stored1 == stored2 {
		t.Fatal("expected unique keys for same original name")
	}
	if orig1 != "informe.pdf" {
		t.Fatalf("original %q", orig1)
	}
	if !strings.HasPrefix(key1, "uploads/") {
		t.Fatalf("key %q", key1)
	}
	if !strings.HasSuffix(stored1, "-informe.pdf") {
		t.Fatalf("stored %q", stored1)
	}
}

func TestBuildUniqueUploadKeyRejectsTraversal(t *testing.T) {
	_, _, _, err := BuildUniqueUploadKey("../../../etc/passwd", "")
	if !IsValidationError(err) {
		t.Fatalf("expected validation error, got %v", err)
	}
}
