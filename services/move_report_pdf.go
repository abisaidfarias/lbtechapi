package services

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

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
	ctx, cancelTO := context.WithTimeout(ctx, 90*time.Second)
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
