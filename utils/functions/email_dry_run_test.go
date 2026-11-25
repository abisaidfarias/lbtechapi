package functions

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// With EMAIL_DRY_RUN=1 nothing goes to SMTP (the empty SMTP config would fail)
// and every email is written to EMAIL_DRY_RUN_DIR with recipients and subject.
func TestEmailDryRunWritesInsteadOfSending(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("EMAIL_DRY_RUN", "1")
	t.Setenv("EMAIL_DRY_RUN_DIR", dir)

	if err := sendBrandedHTMLEmail([]string{"a@lb.test", "b@lb.test"}, "Subject: Hello SMR",
		[]byte("<html><a href=\"https://x/multibanda?validate=1\">Validate request</a></html>"), nil, ""); err != nil {
		t.Fatalf("dry run must not fail: %v", err)
	}
	var body bytes.Buffer
	body.WriteString("Subject: Plain notification \nMIME-version: 1.0;\n\n<p>hi</p>")
	SendNotifications([]string{"c@lb.test"}, body)

	files, _ := filepath.Glob(filepath.Join(dir, "*.eml.html"))
	if len(files) != 2 {
		t.Fatalf("files = %v", files)
	}
	all := ""
	for _, f := range files {
		b, _ := os.ReadFile(f)
		all += string(b)
	}
	for _, want := range []string{"DRY RUN", "To: a@lb.test, b@lb.test", "Subject: Hello SMR",
		"multibanda?validate=1", "To: c@lb.test", "Subject: Plain notification"} {
		if !strings.Contains(all, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestEmailDryRunIsOptIn(t *testing.T) {
	t.Setenv("EMAIL_DRY_RUN", "")
	if EmailDryRun() {
		t.Fatal("dry run must be off unless EMAIL_DRY_RUN is set")
	}
	for _, v := range []string{"1", "true", "YES", "on"} {
		t.Setenv("EMAIL_DRY_RUN", v)
		if !EmailDryRun() {
			t.Errorf("EMAIL_DRY_RUN=%q should enable it", v)
		}
	}
}
