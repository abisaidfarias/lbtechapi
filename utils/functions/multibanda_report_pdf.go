package functions

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/go-pdf/fpdf"
)

// Automatic Multi-Band Certification Report renderer.
//
// Built natively with fpdf (no headless browser) for the same reasons as the
// OABI certificate: predictable, fast and identical across local, dev and prod.
// Unlike the certificate this document is multi-page and data-driven, so it
// uses a repeating header/footer and re-prints table headers across page
// breaks, as the functional spec requires.

const (
	mbrMarginX    = 12.0
	mbrMarginTop  = 26.0
	mbrMarginBot  = 16.0
	mbrPageW      = 210.0
	mbrContentW   = mbrPageW - 2*mbrMarginX
	mbrLineH      = 6.0
	mbrSectionGap = 4.0
)

var (
	mbrHeaderBG  = rgb{11, 23, 54}   // navy
	mbrAccent    = rgb{141, 179, 25} // lime
	mbrSectionBG = rgb{11, 23, 54}
	mbrSubBG     = rgb{236, 253, 245} // light green
	mbrTableHead = rgb{238, 240, 243}
	mbrZebra     = rgb{250, 250, 250}
	mbrBorder    = rgb{229, 231, 235}
	mbrInk       = rgb{17, 24, 39}
	mbrMuted     = rgb{107, 114, 128}
	mbrFailBG    = rgb{254, 226, 226}
	mbrFailInk   = rgb{153, 27, 27}
	mbrOkInk     = rgb{22, 101, 52}
	mbrWhite     = rgb{255, 255, 255}
)

// BuildMultibandaReportPDF renders the report to PDF bytes.
func BuildMultibandaReportPDF(data MultibandaReportPDFData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(mbrMarginX, mbrMarginTop, mbrMarginX)
	pdf.SetAutoPageBreak(true, mbrMarginBot)
	pdf.AliasNbPages("{nb}")

	registerReportImages(pdf, data)

	pdf.SetHeaderFunc(func() { drawReportHeader(pdf, data) })
	pdf.SetFooterFunc(func() { drawReportFooter(pdf, data) })

	pdf.AddPage()

	drawDeviceInformation(pdf, data)

	if data.IncludesSimlock {
		drawSimlockBlock(pdf, data)
	}
	if data.IncludesMultiband {
		drawMultibandBlock(pdf, data)
	}
	drawSAEBlock(pdf, data)
	drawSAEEvidence(pdf, data)
	drawFMRadioBlock(pdf, data)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	if pdf.Err() {
		return nil, fmt.Errorf("build multibanda report pdf: %v", pdf.Error())
	}
	return buf.Bytes(), nil
}

func registerReportImages(pdf *fpdf.Fpdf, data MultibandaReportPDFData) {
	opt := fpdf.ImageOptions{ImageType: "PNG", ReadDpi: false}
	if len(utils.LBOneTrackLogoPNG) > 0 {
		pdf.RegisterImageOptionsReader("mbr_logo", opt, bytes.NewReader(utils.LBOneTrackLogoPNG))
	}
	if len(data.StampImage) > 0 {
		registerImageAutoType(pdf, "mbr_stamp", data.StampImage)
	}
	if len(data.SWVersionPhoto) > 0 {
		registerImageAutoType(pdf, "mbr_sw_version", data.SWVersionPhoto)
	}
	for i, ev := range data.SAEEvidence {
		if len(ev.Image) > 0 {
			registerImageAutoType(pdf, fmt.Sprintf("mbr_evidence_%d", i), ev.Image)
		}
	}
}

// registerImageAutoType sniffs the format so callers can pass either a PNG or
// a JPEG straight from storage without tracking the content type.
func registerImageAutoType(pdf *fpdf.Fpdf, name string, raw []byte) {
	imgType := "PNG"
	if len(raw) > 3 && raw[0] == 0xFF && raw[1] == 0xD8 && raw[2] == 0xFF {
		imgType = "JPG"
	}
	pdf.RegisterImageOptionsReader(name, fpdf.ImageOptions{ImageType: imgType, ReadDpi: false}, bytes.NewReader(raw))
	// A corrupt or unsupported image must not abort the whole report.
	if pdf.Err() {
		pdf.ClearError()
	}
}

func drawReportHeader(pdf *fpdf.Fpdf, data MultibandaReportPDFData) {
	fill(pdf, mbrHeaderBG)
	pdf.Rect(0, 0, mbrPageW, 20, "F")

	if len(utils.LBOneTrackLogoPNG) > 0 {
		info := pdf.GetImageInfo("mbr_logo")
		if info != nil && info.Height() > 0 {
			h := 8.0
			w := h * info.Width() / info.Height()
			if w > 46 {
				w = 46
				h = w * info.Height() / info.Width()
			}
			pdf.ImageOptions("mbr_logo", mbrMarginX, (20-h)/2, w, h, false,
				fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		}
	}

	textColor(pdf, mbrWhite)
	setFont(pdf, "B", 12)
	pdf.SetXY(mbrMarginX, 5)
	pdf.CellFormat(mbrContentW, 5, tr(pdf, "Device Evaluation Report"), "", 0, "R", false, 0, "")

	textColor(pdf, rgb{203, 213, 225})
	setFont(pdf, "", 8)
	pdf.SetXY(mbrMarginX, 11)
	subtitle := data.ProcessTypeLabel
	if data.ReportDate != "" {
		subtitle = subtitle + "  ·  " + data.ReportDate
	}
	pdf.CellFormat(mbrContentW, 5, tr(pdf, subtitle), "", 0, "R", false, 0, "")

	fill(pdf, mbrAccent)
	pdf.Rect(0, 20, mbrPageW, 1.6, "F")

	pdf.SetY(mbrMarginTop)
}

func drawReportFooter(pdf *fpdf.Fpdf, data MultibandaReportPDFData) {
	pdf.SetY(-12)
	draw(pdf, mbrBorder)
	pdf.SetLineWidth(0.2)
	pdf.Line(mbrMarginX, pdf.GetY(), mbrPageW-mbrMarginX, pdf.GetY())

	textColor(pdf, mbrMuted)
	setFont(pdf, "", 7.5)
	pdf.SetY(-9)
	pdf.CellFormat(mbrContentW/2, 5,
		tr(pdf, fmt.Sprintf("© %d LB Technology", data.Year)), "", 0, "L", false, 0, "")
	pdf.CellFormat(mbrContentW/2, 5,
		tr(pdf, fmt.Sprintf("Page %d of {nb}", pdf.PageNo())), "", 0, "R", false, 0, "")
}

// ---- Section helpers ----

// drawSectionTitle starts a section. followingContent is the height the first
// rows of the section need, so a heading never sits alone at the foot of a page
// with its table pushed to the next one.
func drawSectionTitle(pdf *fpdf.Fpdf, title string, followingContent float64) {
	ensureSpace(pdf, 16+followingContent)
	pdf.Ln(mbrSectionGap)
	fill(pdf, mbrSectionBG)
	pdf.Rect(mbrMarginX, pdf.GetY(), mbrContentW, 7, "F")
	textColor(pdf, mbrWhite)
	setFont(pdf, "B", 9.5)
	pdf.SetXY(mbrMarginX+3, pdf.GetY()+1.2)
	pdf.CellFormat(mbrContentW-6, 5, tr(pdf, strings.ToUpper(title)), "", 0, "L", false, 0, "")
	pdf.SetY(pdf.GetY() + 6.5)
	pdf.Ln(2)
}

func drawSubTitle(pdf *fpdf.Fpdf, title string) {
	ensureSpace(pdf, 14)
	fill(pdf, mbrSubBG)
	pdf.Rect(mbrMarginX, pdf.GetY(), mbrContentW, 6, "F")
	textColor(pdf, rgb{6, 95, 70})
	setFont(pdf, "B", 8.5)
	pdf.SetXY(mbrMarginX+3, pdf.GetY()+0.9)
	pdf.CellFormat(mbrContentW-6, 4.2, tr(pdf, title), "", 0, "L", false, 0, "")
	pdf.SetY(pdf.GetY() + 5.6)
	pdf.Ln(1.5)
}

// ensureSpace starts a new page when the remaining height cannot fit the next
// block, so a heading never lands alone at the bottom of a page.
func ensureSpace(pdf *fpdf.Fpdf, needed float64) {
	if pdf.GetY()+needed > 297-mbrMarginBot {
		pdf.AddPage()
	}
}

// ---- Device information ----

func drawDeviceInformation(pdf *fpdf.Fpdf, data MultibandaReportPDFData) {
	drawSectionTitle(pdf, "Device Information", 30)

	pairs := [][2]string{
		{"Manufacturer", data.Manufacturer},
		{"Commercial Name", data.CommercialName},
		{"Technical Model", data.TechnicalModel},
		{"HW Version", data.HardwareVersion},
		{"SW Version", data.SoftwareVersion},
		{"SAR Value", data.SARValue},
		{"CBS Package", data.CBSPackage},
		{"Google Play System Update", data.GooglePlaySystemUpdate},
		{"Operative System Type", data.OperativeSystemType},
		{"Preferred Network", data.PreferredNetwork},
		{"FM Radio", data.FMRadio},
		{"Test Date", data.TestDate},
		{"IMEI", data.IMEI},
		{"S/N", data.SerialNumber},
	}

	// Two label/value pairs per row.
	colW := mbrContentW / 2
	labelW := 38.0
	valueW := colW - labelW - 2

	for i := 0; i < len(pairs); i += 2 {
		ensureSpace(pdf, mbrLineH+2)
		y := pdf.GetY()
		if (i/2)%2 == 1 {
			fill(pdf, mbrZebra)
			pdf.Rect(mbrMarginX, y, mbrContentW, mbrLineH, "F")
		}

		drawFieldPair(pdf, mbrMarginX, y, labelW, valueW, pairs[i][0], pairs[i][1])
		if i+1 < len(pairs) {
			drawFieldPair(pdf, mbrMarginX+colW, y, labelW, valueW, pairs[i+1][0], pairs[i+1][1])
		}
		pdf.SetY(y + mbrLineH)
	}

	drawStampAndPhoto(pdf, data)
}

func drawFieldPair(pdf *fpdf.Fpdf, x, y, labelW, valueW float64, label, value string) {
	if strings.TrimSpace(value) == "" {
		value = "—"
	}
	textColor(pdf, mbrMuted)
	setFont(pdf, "B", 7.8)
	pdf.SetXY(x+1, y+1.4)
	pdf.CellFormat(labelW, 3.6, tr(pdf, label), "", 0, "L", false, 0, "")

	textColor(pdf, mbrInk)
	setFont(pdf, "", 8.6)
	pdf.SetXY(x+labelW+1, y+1.2)
	pdf.CellFormat(valueW, 4, tr(pdf, fitText(pdf, value, valueW-1, 8.6)), "", 0, "L", false, 0, "")
}

func drawStampAndPhoto(pdf *fpdf.Fpdf, data MultibandaReportPDFData) {
	drawCarriersTested(pdf, data)

	hasStamp := len(data.StampImage) > 0
	hasPhoto := len(data.SWVersionPhoto) > 0
	if !hasStamp && !hasPhoto {
		return
	}

	pdf.Ln(3)
	ensureSpace(pdf, 62)
	y := pdf.GetY()
	colW := mbrContentW / 2

	if hasStamp {
		drawImageBox(pdf, mbrMarginX, y, colW-3, 42, "mbr_stamp", "Stamp Type Defined", data.StampLabel)
	}
	if hasPhoto {
		// Taller than the stamp: this is a portrait phone screenshot.
		drawImageBox(pdf, mbrMarginX+colW+3, y, colW-3, 56, "mbr_sw_version", "SW Version", "")
	}

	bottom := y + 44
	if hasPhoto && y+58 > bottom {
		bottom = y + 58
	}
	pdf.SetY(bottom)
}

// drawCarriersTested lists the carriers the device was evaluated against, the
// equivalent of the "CARRIERS TESTED" panel in the previous report.
func drawCarriersTested(pdf *fpdf.Fpdf, data MultibandaReportPDFData) {
	if len(data.CarriersTested) == 0 {
		return
	}

	pdf.Ln(2)
	ensureSpace(pdf, 14)
	textColor(pdf, mbrMuted)
	setFont(pdf, "B", 7.8)
	pdf.SetX(mbrMarginX)
	pdf.CellFormat(mbrContentW, 4, tr(pdf, "Carriers Tested"), "", 2, "L", false, 0, "")

	y := pdf.GetY() + 0.5
	x := mbrMarginX
	setFont(pdf, "", 8.6)
	for _, carrier := range data.CarriersTested {
		label := tr(pdf, carrier)
		w := pdf.GetStringWidth(label) + 9

		// Wrap onto the next line when the row is full.
		if x+w > mbrMarginX+mbrContentW {
			x = mbrMarginX
			y += 6.5
			ensureSpace(pdf, 8)
		}

		fill(pdf, mbrSubBG)
		pdf.Rect(x, y, w, 5.4, "F")

		// The core PDF fonts have no check glyph, so draw one.
		drawCheckMark(pdf, x+3, y+2.7, mbrOkInk)

		textColor(pdf, mbrInk)
		setFont(pdf, "", 8.6)
		pdf.SetXY(x+6, y+0.7)
		pdf.CellFormat(w-8, 4, label, "", 0, "L", false, 0, "")
		x += w + 2
	}
	pdf.SetY(y + 7)
}

// drawImageBox renders a captioned, aspect-preserving image inside a bordered
// area, centered both ways.
func drawImageBox(pdf *fpdf.Fpdf, x, y, w, h float64, imageName, caption, sublabel string) {
	textColor(pdf, mbrMuted)
	setFont(pdf, "B", 7.8)
	pdf.SetXY(x, y)
	pdf.CellFormat(w, 4, tr(pdf, caption), "", 0, "L", false, 0, "")

	boxY := y + 4.5
	boxH := h - 4.5
	draw(pdf, mbrBorder)
	pdf.SetLineWidth(0.2)
	pdf.Rect(x, boxY, w, boxH, "D")

	info := pdf.GetImageInfo(imageName)
	if info != nil && info.Width() > 0 && info.Height() > 0 {
		maxW := w - 4
		maxH := boxH - 4
		if sublabel != "" {
			maxH -= 4
		}
		iw, ih := info.Width(), info.Height()
		scale := maxW / iw
		if h := maxH / ih; h < scale {
			scale = h
		}
		dw, dh := iw*scale, ih*scale
		pdf.ImageOptions(imageName, x+(w-dw)/2, boxY+2+(maxH-dh)/2, dw, dh, false,
			fpdf.ImageOptions{}, 0, "")
	}

	if sublabel != "" {
		textColor(pdf, mbrInk)
		setFont(pdf, "", 7.5)
		pdf.SetXY(x, boxY+boxH-4.5)
		pdf.CellFormat(w, 4, tr(pdf, sublabel), "", 0, "C", false, 0, "")
	}
}

// ---- SIMLOCK ----

func drawSimlockBlock(pdf *fpdf.Fpdf, data MultibandaReportPDFData) {
	drawSectionTitle(pdf, "Detailed Results by Block — SIMLOCK", 22)

	widths := []float64{20, 58, 44, 20, mbrContentW - 142}
	headers := []string{"Test ID", "Test Name", "Carrier", "Result", "Comment"}
	aligns := []string{"L", "L", "L", "C", "L"}

	drawTableHeader(pdf, headers, widths, aligns)
	for i, row := range data.SimlockRows {
		if pdf.GetY()+16 > 297-mbrMarginBot {
			pdf.AddPage()
			drawTableHeader(pdf, headers, widths, aligns)
		}
		cells := []string{row.TestID, row.Name, row.Carrier, row.Result, dashIfEmpty(row.Comment)}
		drawTableRow(pdf, cells, widths, aligns, i%2 == 1, resultTint(row.Result))
	}
}

// ---- Multi-band ----

func drawMultibandBlock(pdf *fpdf.Fpdf, data MultibandaReportPDFData) {
	drawSectionTitle(pdf, "Detailed Results by Block — Multi-band", 40)

	textColor(pdf, mbrMuted)
	setFont(pdf, "", 7.8)
	pdf.SetX(mbrMarginX)
	pdf.MultiCell(mbrContentW, 4,
		tr(pdf, "Each band result consolidates Register, Data and Voice. For 5G NR, Voice includes VoNR."),
		"", "L", false)
	pdf.Ln(1)

	widths := []float64{40, 30, mbrContentW - 70}
	headers := []string{"Band", "Result", "Comment"}
	aligns := []string{"L", "C", "L"}

	for _, group := range data.BandGroups {
		ensureSpace(pdf, 22)
		drawSubTitle(pdf, group.Technology)
		drawTableHeader(pdf, headers, widths, aligns)
		for i, band := range group.Bands {
			if pdf.GetY()+16 > 297-mbrMarginBot {
				pdf.AddPage()
				drawSubTitle(pdf, group.Technology+" (continued)")
				drawTableHeader(pdf, headers, widths, aligns)
			}
			cells := []string{band.Band, band.Result, dashIfEmpty(band.Comment)}
			drawTableRow(pdf, cells, widths, aligns, i%2 == 1, resultTint(band.Result))
		}
		pdf.Ln(2)
	}
}

// ---- SAE ----

func drawSAEBlock(pdf *fpdf.Fpdf, data MultibandaReportPDFData) {
	drawSectionTitle(pdf, "Detailed Results by Block — SAE", 40)

	if data.SAEScenario != "" {
		textColor(pdf, mbrMuted)
		setFont(pdf, "", 7.8)
		pdf.SetX(mbrMarginX)
		pdf.CellFormat(mbrContentW, 4.5,
			tr(pdf, "Applicable SAE Scenario: "+data.SAEScenario), "", 2, "L", false, 0, "")
		pdf.Ln(1.5)
	}

	for _, block := range data.SAEBlocks {
		ensureSpace(pdf, 30)
		drawSubTitle(pdf, fmt.Sprintf("%s — Channel %s", block.ScenarioLabel, block.Channel))

		opCount := float64(len(block.Operators))
		testW := 22.0
		nameW := 52.0
		opW := (mbrContentW - testW - nameW) / opCount

		headers := append([]string{"Test ID", "Test Name"}, block.Operators...)
		widths := append([]float64{testW, nameW}, repeatFloat(opW, len(block.Operators))...)
		aligns := append([]string{"L", "L"}, repeatString("C", len(block.Operators))...)

		drawTableHeader(pdf, headers, widths, aligns)
		for i, row := range block.Rows {
			if pdf.GetY()+16 > 297-mbrMarginBot {
				pdf.AddPage()
				drawSubTitle(pdf, fmt.Sprintf("%s — Channel %s (continued)", block.ScenarioLabel, block.Channel))
				drawTableHeader(pdf, headers, widths, aligns)
			}
			cells := append([]string{row.TestID, row.Name}, row.Results...)
			drawSAETableRow(pdf, cells, widths, aligns, i%2 == 1, len(block.Operators))
		}

		drawSAEComments(pdf, block)
		pdf.Ln(2)
	}
}

// drawSAEComments lists only the failing cells, keeping the matrix itself
// compact while still showing every mandatory justification.
func drawSAEComments(pdf *fpdf.Fpdf, block MultibandaReportSAEBlock) {
	type failure struct{ label, comment string }
	failures := []failure{}
	for _, row := range block.Rows {
		for i, res := range row.Results {
			if res != "FAIL" {
				continue
			}
			comment := ""
			if i < len(row.Comments) {
				comment = row.Comments[i]
			}
			failures = append(failures, failure{
				label:   fmt.Sprintf("%s — %s / %s", row.TestID, row.Name, block.Operators[i]),
				comment: dashIfEmpty(comment),
			})
		}
	}
	if len(failures) == 0 {
		return
	}

	pdf.Ln(1)
	textColor(pdf, mbrFailInk)
	setFont(pdf, "B", 7.8)
	pdf.SetX(mbrMarginX)
	pdf.CellFormat(mbrContentW, 4.5, tr(pdf, "Failure comments"), "", 2, "L", false, 0, "")

	for _, f := range failures {
		ensureSpace(pdf, 8)
		textColor(pdf, mbrInk)
		setFont(pdf, "B", 7.5)
		pdf.SetX(mbrMarginX + 2)
		pdf.CellFormat(70, 4, tr(pdf, f.label), "", 0, "L", false, 0, "")
		textColor(pdf, mbrMuted)
		setFont(pdf, "", 7.5)
		pdf.MultiCell(mbrContentW-74, 4, tr(pdf, f.comment), "", "L", false)
	}
}

func drawSAEEvidence(pdf *fpdf.Fpdf, data MultibandaReportPDFData) {
	if len(data.SAEEvidence) == 0 {
		return
	}
	drawSectionTitle(pdf, "SAE Evidence", 88)

	ensureSpace(pdf, 90)
	y := pdf.GetY()
	colW := mbrContentW / 2
	boxH := 85.0

	for i, ev := range data.SAEEvidence {
		if i >= 2 {
			break
		}
		x := mbrMarginX + float64(i)*colW
		sub := fmt.Sprintf("%s · %s", ev.ScenarioLabel, ev.Operator)
		drawImageBox(pdf, x+float64(i)*1.5, y, colW-3, boxH,
			fmt.Sprintf("mbr_evidence_%d", i), ev.Label, sub)
	}
	pdf.SetY(y + boxH + 3)
}

// ---- FM Radio ----

func drawFMRadioBlock(pdf *fpdf.Fpdf, data MultibandaReportPDFData) {
	drawSectionTitle(pdf, "FM Radio", 20)

	widths := []float64{24, 70, 26, mbrContentW - 120}
	headers := []string{"Test ID", "Test Name", "Result", "Observations"}
	aligns := []string{"L", "L", "C", "L"}

	drawTableHeader(pdf, headers, widths, aligns)

	if !data.FMRadioTestApplies {
		// Declaration only: the conditional test is not enabled.
		note := "The device does not support FM Radio"
		if data.FMRadio == "N/A" {
			note = "FM Radio is not applicable for this device"
		}
		drawTableRow(pdf,
			[]string{"001FM", "FM Radio", data.FMRadio, note},
			widths, aligns, false, rgb{})
		return
	}

	drawTableRow(pdf,
		[]string{"001FM", "FM Radio Application — Non-removable",
			data.FMRadioTestResult, dashIfEmpty(data.FMRadioTestComment)},
		widths, aligns, false, resultTint(data.FMRadioTestResult))
}

// ---- Table primitives ----

func drawTableHeader(pdf *fpdf.Fpdf, headers []string, widths []float64, aligns []string) {
	fill(pdf, mbrTableHead)
	textColor(pdf, mbrHeaderBG)
	setFont(pdf, "B", 7.6)
	y := pdf.GetY()
	x := mbrMarginX
	for i, h := range headers {
		pdf.SetXY(x, y)
		pdf.CellFormat(widths[i], 6.5, tr(pdf, fitText(pdf, h, widths[i]-2, 7.6)), "", 0, aligns[i], true, 0, "")
		x += widths[i]
	}
	pdf.SetY(y + 6.5)
	draw(pdf, mbrBorder)
	pdf.SetLineWidth(0.2)
	pdf.Line(mbrMarginX, pdf.GetY(), mbrMarginX+sumFloat(widths), pdf.GetY())
}

// drawTableRow renders one row. The last column is treated as a comment column
// and wraps onto extra lines, growing the row, so a mandatory failure comment
// is never silently truncated.
func drawTableRow(pdf *fpdf.Fpdf, cells []string, widths []float64, aligns []string, zebra bool, tint rgb) {
	lastIdx := len(cells) - 1
	commentW := widths[lastIdx] - 2

	setFont(pdf, "", 7.6)
	lines := pdf.SplitLines([]byte(tr(pdf, cells[lastIdx])), commentW)
	if len(lines) == 0 {
		lines = [][]byte{[]byte("")}
	}

	rowH := 6.2
	if h := float64(len(lines))*3.8 + 2.4; h > rowH {
		rowH = h
	}

	y := pdf.GetY()
	if tint != (rgb{}) {
		fill(pdf, tint)
		pdf.Rect(mbrMarginX, y, sumFloat(widths), rowH, "F")
	} else if zebra {
		fill(pdf, mbrZebra)
		pdf.Rect(mbrMarginX, y, sumFloat(widths), rowH, "F")
	}

	x := mbrMarginX
	for i, c := range cells {
		if i >= len(widths) {
			break
		}
		textColor(pdf, cellInk(c))
		if isResultToken(c) {
			setFont(pdf, "B", 7.6)
		} else {
			setFont(pdf, "", 7.6)
		}

		if i == lastIdx {
			ly := y + 1.4
			for _, line := range lines {
				pdf.SetXY(x, ly)
				pdf.CellFormat(widths[i], 3.8, string(line), "", 0, aligns[i], false, 0, "")
				ly += 3.8
			}
		} else {
			pdf.SetXY(x, y+(rowH-4.2)/2)
			pdf.CellFormat(widths[i], 4.2, tr(pdf, fitText(pdf, c, widths[i]-2, 7.6)), "", 0, aligns[i], false, 0, "")
		}
		x += widths[i]
	}

	pdf.SetY(y + rowH)
	draw(pdf, mbrBorder)
	pdf.SetLineWidth(0.15)
	pdf.Line(mbrMarginX, pdf.GetY(), mbrMarginX+sumFloat(widths), pdf.GetY())
}

// drawSAETableRow tints individual result cells instead of the whole row, so a
// single failing operator stands out without flagging the other three.
func drawSAETableRow(pdf *fpdf.Fpdf, cells []string, widths []float64, aligns []string, zebra bool, opCount int) {
	y := pdf.GetY()
	rowH := 6.2

	if zebra {
		fill(pdf, mbrZebra)
		pdf.Rect(mbrMarginX, y, sumFloat(widths), rowH, "F")
	}

	firstOp := len(cells) - opCount
	x := mbrMarginX
	for i, c := range cells {
		if i >= len(widths) {
			break
		}
		if i >= firstOp {
			if tint := resultTint(c); tint != (rgb{}) {
				fill(pdf, tint)
				pdf.Rect(x, y, widths[i], rowH, "F")
			}
		}
		textColor(pdf, cellInk(c))
		if isResultToken(c) {
			setFont(pdf, "B", 7.6)
		} else {
			setFont(pdf, "", 7.6)
		}
		pdf.SetXY(x, y+1)
		pdf.CellFormat(widths[i], 4.2, tr(pdf, fitText(pdf, c, widths[i]-2, 7.6)), "", 0, aligns[i], false, 0, "")
		x += widths[i]
	}

	pdf.SetY(y + rowH)
	draw(pdf, mbrBorder)
	pdf.SetLineWidth(0.15)
	pdf.Line(mbrMarginX, pdf.GetY(), mbrMarginX+sumFloat(widths), pdf.GetY())
}

// drawCheckMark strokes a tick centred on (cx, cy). The core PDF fonts carry no
// check glyph, so it is drawn rather than typed.
func drawCheckMark(pdf *fpdf.Fpdf, cx, cy float64, color rgb) {
	draw(pdf, color)
	pdf.SetLineWidth(0.45)
	pdf.Line(cx-1.4, cy, cx-0.4, cy+1.1)
	pdf.Line(cx-0.4, cy+1.1, cx+1.5, cy-1.3)
}

// ---- Small helpers ----

// resultTint gives failing/unsupported results a background so they are
// findable at a glance. Passing results stay neutral to avoid a wall of green.
func resultTint(result string) rgb {
	switch strings.TrimSpace(result) {
	case "FAIL", "NO OK":
		return mbrFailBG
	case "NOT SUPPORTED", "N/A":
		return rgb{243, 244, 246}
	default:
		return rgb{}
	}
}

func cellInk(value string) rgb {
	switch strings.TrimSpace(value) {
	case "FAIL", "NO OK":
		return mbrFailInk
	case "PASS", "OK":
		return mbrOkInk
	case "NOT SUPPORTED", "N/A":
		return mbrMuted
	default:
		return mbrInk
	}
}

func isResultToken(value string) bool {
	switch strings.TrimSpace(value) {
	case "PASS", "FAIL", "OK", "NO OK", "NOT SUPPORTED", "N/A":
		return true
	}
	return false
}

func dashIfEmpty(v string) string {
	if strings.TrimSpace(v) == "" {
		return "—"
	}
	return v
}

// fitText shrinks nothing but truncates with an ellipsis so a long value can
// never bleed into the neighbouring column.
func fitText(pdf *fpdf.Fpdf, text string, maxW, size float64) string {
	if strings.TrimSpace(text) == "" {
		return text
	}
	if pdf.GetStringWidth(tr(pdf, text)) <= maxW {
		return text
	}
	truncated := text
	for len(truncated) > 1 && pdf.GetStringWidth(tr(pdf, truncated+"...")) > maxW {
		truncated = truncated[:len(truncated)-1]
	}
	return truncated + "..."
}

func sumFloat(values []float64) float64 {
	total := 0.0
	for _, v := range values {
		total += v
	}
	return total
}

func repeatFloat(v float64, n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = v
	}
	return out
}

func repeatString(v string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = v
	}
	return out
}
