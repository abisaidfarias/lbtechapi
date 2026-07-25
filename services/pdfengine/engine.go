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
	// A render on a small EB instance regularly needs several minutes for
	// chromedp + heavy HTML (see moveReportPDFTimeout). A local-sized budget
	// here makes the job fail on dev/prod while passing on a dev machine.
	defaultRenderTimeoutSec = 300
)

// RenderTimeout reports the per-render budget so callers can size their own job
// deadlines around it instead of hardcoding a value that silently disagrees.
func RenderTimeout() time.Duration {
	return time.Duration(envIntPositive("SHIPMENT_CERT_PDF_TIMEOUT_SEC", defaultRenderTimeoutSec)) * time.Second
}

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
	return nil
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

	pdf, err := e.renderOnce(parent, html)
	// A cancelled caller looks like a dead browser to isBrowserDead; check the
	// caller first so giving up does not restart Chrome and render again.
	if err != nil && parent.Err() != nil {
		return nil, parent.Err()
	}
	if err != nil && isBrowserDead(err) {
		if rerr := e.restartAllocator(); rerr != nil {
			return nil, rerr
		}
		pdf, err = e.renderOnce(parent, html)
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

func (e *Engine) renderOnce(parent context.Context, html string) ([]byte, error) {
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

	e.mu.Lock()
	allocCtx := e.allocCtx
	e.mu.Unlock()

	jobCtx, cancelJob := chromedp.NewContext(allocCtx)
	defer cancelJob()

	ctx, cancel := context.WithTimeout(jobCtx, RenderTimeout())
	defer cancel()

	// chromedp requires the allocator as its parent, so the caller's context
	// cannot be chained in directly. Watch it instead: without this, a caller
	// that times out leaves Chrome rendering in the background, and on a small
	// instance those orphans pile up and starve every later render.
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-parent.Done():
			cancel()
		case <-done:
		}
	}()

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
