package functions

import (
	"strings"
	"testing"
)

func TestRenderOutlookEmailButtonIncludesOutlookVML(t *testing.T) {
	got := string(RenderOutlookEmailButton("https://example.com/file.pdf", "Excel file", EmailButtonColorSecondary))
	if !strings.Contains(got, "v:roundrect") {
		t.Fatal("expected VML roundrect for Outlook")
	}
	if !strings.Contains(got, `bgcolor="`+EmailButtonColorSecondary+`"`) {
		t.Fatal("expected bgcolor on td for Outlook fallback")
	}
	if !strings.Contains(got, "padding:14px 24px") {
		t.Fatal("expected padded anchor for non-Outlook clients")
	}
}

func TestOutlookButtonVMLWidthScalesWithLabel(t *testing.T) {
	short := outlookButtonVMLWidth("Excel file")
	long := outlookButtonVMLWidth("Download OABI Certificate")
	if long <= short {
		t.Fatalf("expected wider button for longer label, got short=%d long=%d", short, long)
	}
}
