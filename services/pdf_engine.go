package services

import "context"

// PDFEngine renders HTML documents to PDF bytes.
type PDFEngine interface {
	RenderPDF(ctx context.Context, html string) ([]byte, error)
}
