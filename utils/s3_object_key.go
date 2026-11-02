package utils

import (
	"net/url"
	"strings"
)

const defaultS3Bucket = "lbtechimages"

// ObjectKeyFromPublicURL extracts the S3 object key from a public object URL
// served from the lbtechimages bucket.
func ObjectKeyFromPublicURL(fileURL string) (string, bool) {
	fileURL = strings.TrimSpace(fileURL)
	if fileURL == "" {
		return "", false
	}
	u, err := url.Parse(fileURL)
	if err != nil {
		return "", false
	}
	path := strings.TrimPrefix(u.Path, "/")
	if path == "" {
		return "", false
	}

	host := strings.ToLower(u.Host)
	if strings.HasPrefix(host, defaultS3Bucket+".") {
		return path, true
	}
	if host == "s3.amazonaws.com" || strings.HasPrefix(host, "s3.") {
		parts := strings.SplitN(path, "/", 2)
		if len(parts) == 2 && parts[0] == defaultS3Bucket {
			return parts[1], true
		}
	}
	return "", false
}
