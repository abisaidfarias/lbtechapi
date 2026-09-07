package utils

import (
	"embed"
	"os"
	"strings"
)

// Stamp images for the Automatic Multi-Band Certification Report.
//
// These are the official artwork supplied by the project owner. The file name
// states which bands the device DOES support: "2G_4G" is a device that supports
// 2G and 4G but not 3G, and "4G_and_below" supports everything up to 4G.
//
//go:embed stampImages/*
var multibandaStampFS embed.FS

// Extensions tried when resolving a stamp, so replacing the artwork with a
// different image format does not require a code change.
var stampImageExtensions = []string{".jpeg", ".jpg", ".png"}

// MultibandaStampImage returns the image bytes for a catalog stamp image key,
// or nil when the asset is missing so the renderer can skip it gracefully.
func MultibandaStampImage(imageKey string) []byte {
	imageKey = strings.TrimSpace(imageKey)
	if imageKey == "" {
		return nil
	}
	for _, ext := range stampImageExtensions {
		if data, err := multibandaStampFS.ReadFile("stampImages/" + imageKey + ext); err == nil {
			return data
		}
	}
	return nil
}

// MultibandaStampContentType sniffs the media type of stamp bytes so the HTTP
// handler advertises what it actually serves: the artwork format is whatever
// the project owner drops into stampImages, not a fixed PNG.
func MultibandaStampContentType(image []byte) string {
	if len(image) > 2 && image[0] == 0xFF && image[1] == 0xD8 && image[2] == 0xFF {
		return "image/jpeg"
	}
	return "image/png"
}

// EnvOrDefault reads an environment variable, falling back when unset or blank.
func EnvOrDefault(name, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return fallback
}
