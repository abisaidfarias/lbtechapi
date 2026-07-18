package pdfengine

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

const (
	defaultMaxRendersBeforeRestart = 200
	defaultRenderTimeoutSec        = 90
)

type Engine struct {
	mu        sync.Mutex
	sem       chan struct{}
	allocCtx  context.Context
	allocStop context.CancelFunc
	renders   int
}

func New() (*Engine, error) {
	concurrency := envIntPositive("SHIPMENT_CERT_PDF_CONCURRENCY", 1)
	e := &Engine{sem: make(chan struct{}, concurrency)}
	if err := e.startAllocator(); err != nil {
		return nil, err
	}
	return e, nil
}

func chromeAllocatorOptions() []chromedp.ExecAllocatorOption {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	if runtime.GOOS == "linux" {
		opts = append(opts, chromedp.Flag("no-zygote", true))
	}
	if p := strings.TrimSpace(os.Getenv("CHROME_PATH")); p != "" {
		opts = append(opts, chromedp.ExecPath(p))
	}
	return opts
}

func (e *Engine) startAllocator() error {
	opts := chromeAllocatorOptions()
	e.allocCtx, e.allocStop = chromedp.NewExecAllocator(context.Background(), opts...)

	warmupCtx, cancelWarmup := chromedp.NewContext(e.allocCtx)
	defer cancelWarmup()
	warmupTO, cancelTO := context.WithTimeout(warmupCtx, 30*time.Second)
	defer cancelTO()
	return chromedp.Run(warmupTO, chromedp.Navigate("about:blank"))
}

func (e *Engine) restartAllocator() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.allocStop != nil {
		e.allocStop()
	}
	e.renders = 0
	return e.startAllocator()
}

func (e *Engine) RenderPDF(parent context.Context, html string) ([]byte, error) {
	select {
	case e.sem <- struct{}{}:
		defer func() { <-e.sem }()
	case <-parent.Done():
		return nil, parent.Err()
	}

	pdf, err := e.renderOnce(html)
	if err != nil && isBrowserDead(err) {
		if rerr := e.restartAllocator(); rerr != nil {
			return nil, rerr
		}
		pdf, err = e.renderOnce(html)
	}

	e.mu.Lock()
	e.renders++
	needsRestart := e.renders >= defaultMaxRendersBeforeRestart
	e.mu.Unlock()
	if needsRestart {
		go func() {
			_ = e.restartAllocator()
		}()
	}
	return pdf, err
}

func (e *Engine) renderOnce(html string) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "lbtech-shipment-cert-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	htmlPath := filepath.Join(tmpDir, "certificate.html")
	if err := os.WriteFile(htmlPath, []byte(html), 0600); err != nil {
		return nil, err
	}

	fileURL, err := fileURL(htmlPath)
	if err != nil {
		return nil, err
	}

	timeoutSec := envIntPositive("SHIPMENT_CERT_PDF_TIMEOUT_SEC", defaultRenderTimeoutSec)

	e.mu.Lock()
	allocCtx := e.allocCtx
	e.mu.Unlock()

	jobCtx, cancelJob := chromedp.NewContext(allocCtx)
	defer cancelJob()

	ctx, cancel := context.WithTimeout(jobCtx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	var pdf []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate(fileURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdf, _, err = page.PrintToPDF().
				WithPaperWidth(8.27).
				WithPaperHeight(11.69).
				WithMarginTop(0).
				WithMarginBottom(0).
				WithMarginLeft(0).
				WithMarginRight(0).
				WithPrintBackground(true).
				WithPreferCSSPageSize(true).
				Do(ctx)
			return err
		}),
	)
	return pdf, err
}

func fileURL(path string) (string, error) {
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

func isBrowserDead(err error) bool {
	if err == nil {
		return false
	}
	if err == context.Canceled {
		return true
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "context canceled") ||
		strings.Contains(s, "chrome failed to start") ||
		strings.Contains(s, "websocket") ||
		strings.Contains(s, "connection refused") ||
		strings.Contains(s, "target closed")
}

func envIntPositive(name string, defaultValue int) int {
	s := strings.TrimSpace(os.Getenv(name))
	if s == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return defaultValue
	}
	return n
}
