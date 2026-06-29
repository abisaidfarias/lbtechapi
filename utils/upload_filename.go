package utils

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gofrs/uuid"
)

const maxUploadBaseNameLen = 100

var unsafeUploadNameChars = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

// ValidateUploadOriginalFileName rejects path traversal and path-like names.
func ValidateUploadOriginalFileName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return NewValidationError("file name is required")
	}
	if strings.Contains(name, "..") {
		return NewValidationError("invalid file name")
	}
	if strings.ContainsAny(name, `/\`) {
		return NewValidationError("invalid file name")
	}
	if strings.HasPrefix(name, ".") && name != "." && !strings.HasPrefix(name, "..") {
		// hidden files ok unless only dots
	}
	if name == "." || name == ".." {
		return NewValidationError("invalid file name")
	}
	return nil
}

// SanitizeUploadFileName produces a safe file name segment for S3 keys.
func SanitizeUploadFileName(original string, fallbackExt string) string {
	base := strings.TrimSpace(original)
	base = strings.ReplaceAll(base, `\`, `/`)
	base = filepath.Base(base)
	base = strings.TrimSpace(base)

	ext := strings.ToLower(filepath.Ext(base))
	baseName := strings.TrimSuffix(base, filepath.Ext(base))
	baseName = strings.TrimSpace(baseName)
	baseName = strings.ReplaceAll(baseName, " ", "-")
	baseName = unsafeUploadNameChars.ReplaceAllString(baseName, "-")
	baseName = strings.Trim(baseName, "-._")

	if ext == "" && fallbackExt != "" {
		if !strings.HasPrefix(fallbackExt, ".") {
			ext = "." + fallbackExt
		} else {
			ext = fallbackExt
		}
	}
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	if baseName == "" {
		baseName = "document"
	}
	if utf8.RuneCountInString(baseName) > maxUploadBaseNameLen {
		baseName = truncateRunes(baseName, maxUploadBaseNameLen)
	}

	return baseName + ext
}

// BuildUniqueUploadKey returns uploads/{uuid}-{sanitized} and the stored file segment.
func BuildUniqueUploadKey(originalFileName, mimeExt string) (key, storedFileName, displayOriginal string, err error) {
	if err = ValidateUploadOriginalFileName(originalFileName); err != nil {
		return "", "", "", err
	}

	displayOriginal = filepath.Base(strings.ReplaceAll(strings.TrimSpace(originalFileName), `\`, `/`))
	sanitized := SanitizeUploadFileName(displayOriginal, mimeExt)
	id, err := uuid.NewV4()
	if err != nil {
		return "", "", "", fmt.Errorf("generate upload id: %w", err)
	}

	storedFileName = fmt.Sprintf("%s-%s", id.String(), sanitized)
	key = "uploads/" + storedFileName
	return key, storedFileName, displayOriginal, nil
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	var b strings.Builder
	for i, r := range s {
		if i >= max {
			break
		}
		b.WriteRune(r)
	}
	return b.String()
}
