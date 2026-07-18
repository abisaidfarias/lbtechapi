package services

import (
	"context"
	"log"
	"sync"

	"github.com/abisaidfarias/lbtechapi/services/pdfengine"
)

type lazyPDFEngine struct {
	once sync.Once
	eng  PDFEngine
	err  error
}

func NewLazyPDFEngine() PDFEngine {
	return &lazyPDFEngine{}
}

func (l *lazyPDFEngine) RenderPDF(ctx context.Context, html string) ([]byte, error) {
	l.once.Do(func() {
		e, err := pdfengine.New()
		if err != nil {
			l.err = err
			log.Printf("pdf engine init failed: %v", err)
			return
		}
		l.eng = e
	})
	if l.err != nil {
		return nil, l.err
	}
	return l.eng.RenderPDF(ctx, html)
}
