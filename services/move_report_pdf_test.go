package services

import (
	"os"
	"testing"
)

func TestShipmentControlCertificateHTMLToPDFSimple(t *testing.T) {
	if os.Getenv("RUN_PDFENGINE_TEST") == "" {
		t.Skip("set RUN_PDFENGINE_TEST=1 to run")
	}
	html := []byte(`<!DOCTYPE html><html><body><h1>test</h1></body></html>`)
	pdf, err := shipmentControlCertificateHTMLToPDF(html)
	if err != nil {
		t.Fatalf("htmlToPDF: %v", err)
	}
	if len(pdf) < 100 {
		t.Fatalf("pdf too small: %d bytes", len(pdf))
	}
}
