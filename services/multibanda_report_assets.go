package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
)

const (
	reportImageFetchTimeout = 20 * time.Second
	reportImageMaxBytes     = 12 << 20 // a little over the 10 MB upload cap
)

// fetchReportImage downloads an uploaded image so it can be embedded in the
// PDF. A missing or unreachable image must not fail the whole report: the
// renderer simply leaves that box empty, so errors are logged and swallowed.
func fetchReportImage(ctx context.Context, rawURL string) []byte {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil
	}

	reqCtx, cancel := context.WithTimeout(ctx, reportImageFetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, rawURL, nil)
	if err != nil {
		log.Printf("multibanda report image request (%s): %v", rawURL, err)
		return nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("multibanda report image fetch (%s): %v", rawURL, err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("multibanda report image fetch (%s): status %d", rawURL, resp.StatusCode)
		return nil
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, reportImageMaxBytes))
	if err != nil {
		log.Printf("multibanda report image read (%s): %v", rawURL, err)
		return nil
	}
	return data
}

// stampImageBytes resolves the engineer-selected stamp to its embedded asset.
// The five final images are supplied by the project owner and loaded into
// utils.ShipmentStampImages keyed by the catalog's ImageKey.
func stampImageBytes(stampCode string) ([]byte, string) {
	stamp, ok := enums.StampByCode(strings.TrimSpace(stampCode))
	if !ok {
		return nil, ""
	}
	return utils.MultibandaStampImage(stamp.ImageKey), stamp.Label
}

func multibandaReportObjectKey(multibandaID, controlName string) string {
	prefix := strings.TrimSpace(utils.EnvOrDefault("MULTIBANDA_REPORT_S3_PREFIX", "multibanda-reports"))
	prefix = strings.Trim(prefix, "/")
	safe := strings.NewReplacer("/", "-", "\\", "-", " ", "-").Replace(controlName)
	return fmt.Sprintf("%s/%s-%s.pdf", prefix, multibandaID, safe)
}
