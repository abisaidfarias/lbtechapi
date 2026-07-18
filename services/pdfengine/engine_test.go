package pdfengine

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestEngineRenderPDFSimple(t *testing.T) {
	if os.Getenv("RUN_PDFENGINE_TEST") == "" {
		t.Skip("set RUN_PDFENGINE_TEST=1 to run")
	}
	e, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	html := `<!DOCTYPE html><html><body><h1>test</h1></body></html>`
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pdf, err := e.RenderPDF(ctx, html)
	if err != nil {
		t.Fatalf("RenderPDF: %v", err)
	}
	if len(pdf) < 100 {
		t.Fatalf("pdf too small: %d bytes", len(pdf))
	}
}
