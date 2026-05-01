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
	const defaultSec = 300
	s := strings.TrimSpace(os.Getenv("MOVE_REPORT_PDF_TIMEOUT_SEC"))
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
		return "PDF timed out; increase MOVE_REPORT_PDF_TIMEOUT_SEC or instance CPU (email sent without attachment)"
	}
	return "install Chrome/Chromium or set CHROME_PATH (email sent without attachment)"
}

func movementHTMLToPDF(html []byte) ([]byte, error) {
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
	ctx, cancelTO := context.WithTimeout(ctx, moveReportPDFTimeout())
	defer cancelTO()

	var pdf []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate(fileURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdf, _, err = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
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
