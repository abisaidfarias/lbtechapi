package services

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func moveReportPDFTimeout() time.Duration {
	const defaultSec = 600 // 10m; dev/EB often needs more than 5m for chromedp + heavy HTML
	return envDurationSec("MOVE_REPORT_PDF_TIMEOUT_SEC", defaultSec)
}

// moveReportPDFFastTimeout is the per-attempt timeout used in the "fast" phase
// (one shot at producing the PDF before the email goes out). Default is short
// so the email is not held back by a stuck Chrome process; the slow background
// phase uses moveReportPDFTimeout for retries.
func moveReportPDFFastTimeout() time.Duration {
	const defaultSec = 90
	return envDurationSec("MOVE_REPORT_PDF_FAST_TIMEOUT_SEC", defaultSec)
}

func envDurationSec(name string, defaultSec int) time.Duration {
	s := strings.TrimSpace(os.Getenv(name))
	if s == "" {
		return time.Duration(defaultSec) * time.Second
	}
	sec, err := strconv.Atoi(s)
	if err != nil || sec <= 0 {
		return time.Duration(defaultSec) * time.Second
	}
	return time.Duration(sec) * time.Second
}

func moveReportPDFErrHint(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "PDF timed out; increase MOVE_REPORT_PDF_TIMEOUT_SEC / MOVE_REPORT_PDF_FAST_TIMEOUT_SEC or instance CPU (email sent without link this time, retried in background)"
	}
	return "install Chrome/Chromium or set CHROME_PATH (email sent without link this time, retried in background)"
}

func movementHTMLToPDF(html []byte) ([]byte, error) {
	return htmlToPDF(html, moveReportPDFTimeout(), 0, 0)
}

func shipmentControlCertificateHTMLToPDF(html []byte) ([]byte, error) {
	const a4WidthIn = 8.27
	const a4HeightIn = 11.69
	return htmlToPDF(html, moveReportPDFTimeout(), a4WidthIn, a4HeightIn)
}

func movementHTMLToPDFWithTimeout(html []byte, timeout time.Duration) ([]byte, error) {
	if timeout <= 0 {
		timeout = moveReportPDFTimeout()
	}
	return htmlToPDF(html, timeout, 0, 0)
}

func movementHTMLToPDFInternal(html []byte, timeout time.Duration) ([]byte, error) {
	return htmlToPDF(html, timeout, 0, 0)
}

func htmlToPDF(html []byte, timeout time.Duration, paperWidthIn, paperHeightIn float64) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "lbtech-move-report-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	htmlPath := filepath.Join(tmpDir, "report.html")
	if err := os.WriteFile(htmlPath, html, 0600); err != nil {
		return nil, err
	}
	log.Printf("move report: temp HTML for Chrome at %s (folder removed after PDF)", htmlPath)

	fileURL, err := reportFileURL(htmlPath)
	if err != nil {
		return nil, err
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	if p := strings.TrimSpace(os.Getenv("CHROME_PATH")); p != "" {
		opts = append(opts, chromedp.ExecPath(p))
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()
	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()
	ctx, cancelTO := context.WithTimeout(ctx, timeout)
	defer cancelTO()

	var pdf []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate(fileURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfAction := page.PrintToPDF().WithPrintBackground(true)
			if paperWidthIn > 0 && paperHeightIn > 0 {
				pdfAction = pdfAction.
					WithPaperWidth(paperWidthIn).
					WithPaperHeight(paperHeightIn).
					WithPreferCSSPageSize(true).
					WithMarginTop(0).
					WithMarginBottom(0).
					WithMarginLeft(0).
					WithMarginRight(0)
			}
			pdf, _, err = pdfAction.Do(ctx)
			return err
		}),
	)
	return pdf, err
}

func reportFileURL(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	slash := filepath.ToSlash(abs)
	if runtime.GOOS == "windows" {
		return "file:///" + slash, nil
	}
	return "file://" + slash, nil
}
