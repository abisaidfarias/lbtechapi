package functions

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
)

// Email dry run (opt-in, for local tests against shared data): with
// EMAIL_DRY_RUN=1 nothing is sent over SMTP. Each email is logged (recipients
// and subject) and written whole to EMAIL_DRY_RUN_DIR (default
// <temp>/lbtech-email-dryrun) so it can be opened in a browser.

var dryRunSeq atomic.Int64

// EmailDryRun reports whether EMAIL_DRY_RUN is enabled.
func EmailDryRun() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("EMAIL_DRY_RUN"))) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

// EmailDryRunDir is where dry-run emails are written.
func EmailDryRunDir() string {
	if dir := strings.TrimSpace(os.Getenv("EMAIL_DRY_RUN_DIR")); dir != "" {
		return dir
	}
	return filepath.Join(os.TempDir(), "lbtech-email-dryrun")
}

var subjectLine = regexp.MustCompile(`(?m)^Subject:\s*(.*?)\r?$`)

// subjectFromMessage reads the Subject header of a raw message.
func subjectFromMessage(msg []byte) string {
	if m := subjectLine.FindSubmatch(msg); m != nil {
		return strings.TrimSpace(string(m[1]))
	}
	return ""
}

// recordDryRunEmail logs the email instead of sending it and returns the
// file it was written to.
func recordDryRunEmail(toList []string, subject string, msg []byte) string {
	subject = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(subject), "Subject:"))
	dir := EmailDryRunDir()
	path := ""
	if err := os.MkdirAll(dir, 0o755); err == nil {
		name := fmt.Sprintf("%s_%03d.eml.html", time.Now().Format("20060102_150405"), dryRunSeq.Add(1))
		path = filepath.Join(dir, name)
		header := fmt.Sprintf("<!--\nDRY RUN — not sent\nTo: %s\nSubject: %s\n-->\n",
			strings.Join(toList, ", "), subject)
		if err := os.WriteFile(path, append([]byte(header), msg...), 0o644); err != nil {
			path = ""
		}
	}
	log.Printf("EMAIL_DRY_RUN not sent: to=%v subject=%q file=%s", toList, subject, path)
	return path
}
