package mapping_test

import (
	"testing"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
)

const storedReportURL = "https://s3/report.pdf?v=1757000000"

// A client that does not send the field must not clear it. This is what keeps
// the automatic report's link alive: the server writes it at generation time,
// so any client holding a copy loaded before that echoes an empty string back.
func TestPreserveTestReportURLKeepsStoredValueWhenRequestIsEmpty(t *testing.T) {
	multibanda := &models.Multibanda{}

	mapping.PreserveMultibandaTestReportURL(multibanda, storedReportURL)

	if multibanda.TestReportUrl != storedReportURL {
		t.Errorf("got %q, want the stored url %q", multibanda.TestReportUrl, storedReportURL)
	}
}

// Whitespace is not a value: a form that submits a blank field must behave the
// same as one that omits it.
func TestPreserveTestReportURLTreatsBlankAsAbsent(t *testing.T) {
	multibanda := &models.Multibanda{TestReportUrl: "   "}

	mapping.PreserveMultibandaTestReportURL(multibanda, storedReportURL)

	if multibanda.TestReportUrl != storedReportURL {
		t.Errorf("got %q, want the stored url %q", multibanda.TestReportUrl, storedReportURL)
	}
}

// The manual upload still has to win: it sends a real URL and that must land.
func TestPreserveTestReportURLLetsANewValueThrough(t *testing.T) {
	const uploaded = "https://s3/manual-upload.pdf"
	multibanda := &models.Multibanda{TestReportUrl: uploaded}

	mapping.PreserveMultibandaTestReportURL(multibanda, storedReportURL)

	if multibanda.TestReportUrl != uploaded {
		t.Errorf("got %q, want the uploaded url %q", multibanda.TestReportUrl, uploaded)
	}
}

// Nothing stored and nothing sent stays empty rather than inventing a value.
func TestPreserveTestReportURLLeavesEmptyWhenThereIsNothingToKeep(t *testing.T) {
	multibanda := &models.Multibanda{}

	mapping.PreserveMultibandaTestReportURL(multibanda, "")

	if multibanda.TestReportUrl != "" {
		t.Errorf("got %q, want an empty url", multibanda.TestReportUrl)
	}
}

func TestPreserveTestReportURLIgnoresNil(t *testing.T) {
	// Must not panic.
	mapping.PreserveMultibandaTestReportURL(nil, storedReportURL)
}
