package utils

import (
	"embed"
	"os"
	"strings"
)

// Stamp images for the Automatic Multi-Band Certification Report.
//
// PLACEHOLDERS: these approximate the supplied design so the whole flow works
// end to end. The five final images come from the project owner — replacing the
// files in utils/stampImages/ (same names) swaps them in with no code change.
//
//go:embed stampImages/*.png
var multibandaStampFS embed.FS

// MultibandaStampImage returns the PNG bytes for a catalog stamp image key, or
// nil when the asset is missing so the renderer can skip it gracefully.
func MultibandaStampImage(imageKey string) []byte {
	imageKey = strings.TrimSpace(imageKey)
	if imageKey == "" {
		return nil
	}
	data, err := multibandaStampFS.ReadFile("stampImages/" + imageKey + ".png")
	if err != nil {
		return nil
	}
	return data
}

// EnvOrDefault reads an environment variable, falling back when unset or blank.
func EnvOrDefault(name, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return fallback
}
