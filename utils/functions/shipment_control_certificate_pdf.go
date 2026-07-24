package functions

import (
	"bytes"
	"fmt"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/go-pdf/fpdf"
)

// Certificate layout is a fixed single-page A4 document, so it is drawn
// natively with fpdf instead of rendering HTML through a headless browser.
// This removes Chromium from the request path entirely: rendering is a few
// milliseconds and a few MB of RAM, and behaves identically on local, dev and
// prod regardless of instance size.

// Palette mirrors the HTML template (utils/htmlMessageTemplate/shipment_control_certificate.html).
type rgb struct{ r, g, b int }

var (
	certNavy      = rgb{11, 23, 54}    // #0b1736
	certAccent    = rgb{197, 217, 38}  // #c5d926
	certPageBG    = rgb{247, 247, 242}  // #f7f7f2
	certInk       = rgb{17, 24, 39}    // #111827
	certBody      = rgb{55, 65, 81}    // #374151
	certMuted     = rgb{107, 114, 128} // #6b7280
	certBorder    = rgb{229, 231, 235} // #e5e7eb
	certTableHead = rgb{238, 240, 243} // #eef0f3
	certLightText = rgb{203, 213, 225} // #cbd5e1
)

// certificate sheet geometry in millimeters.
const (
	certSheetX0   = 12.0
	certSheetX1   = 198.0
	certPad       = 8.5
	certContentX0 = certSheetX0 + certPad
	certContentX1 = certSheetX1 - certPad
	certContentW  = certContentX1 - certContentX0
)

// BuildShipmentControlCertificatePDF renders the certificate to PDF bytes
// natively, without a browser.
func BuildShipmentControlCertificatePDF(data ShipmentControlCertificateData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	// Page background.
	fill(pdf, certPageBG)
	pdf.Rect(0, 0, 210, 297, "F")

	// White sheet with border.
	fill(pdf, rgb{255, 255, 255})
	draw(pdf, certBorder)
	pdf.SetLineWidth(0.2)
	pdf.Rect(certSheetX0, 12, certSheetX1-certSheetX0, 273, "FD")

	if err := registerCertImages(pdf, data); err != nil {
		return nil, err
	}

	y := 12.0
	y = drawCertHeader(pdf, data, y)
	y = drawAccentDivider(pdf, y)
	y = drawCompanies(pdf, data, y)
	y = drawInnerDivider(pdf, y+2)
	y = drawDeviceData(pdf, data, y+3)
	y = drawInnerDivider(pdf, y+2)
	y = drawGuides(pdf, data, y+3)
	y = drawInnerDivider(pdf, y+2)
	y = drawChecklist(pdf, y+3)
	y = drawInnerDivider(pdf, y+2)
	drawLegalAndSignature(pdf, data, y+3)

	drawFooter(pdf, data)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	if pdf.Err() {
		return nil, fmt.Errorf("build certificate pdf: %v", pdf.Error())
	}
	return buf.Bytes(), nil
}

func registerCertImages(pdf *fpdf.Fpdf, data ShipmentControlCertificateData) error {
	if len(utils.LBOneTrackLogoPNG) > 0 {
		opt := fpdf.ImageOptions{ImageType: "PNG", ReadDpi: false}
		pdf.RegisterImageOptionsReader("lb_logo", opt, bytes.NewReader(utils.LBOneTrackLogoPNG))
	}
	if len(utils.ShipmentControlOfficialSealPNG) > 0 {
		opt := fpdf.ImageOptions{ImageType: "PNG", ReadDpi: false}
		pdf.RegisterImageOptionsReader("official_seal", opt, bytes.NewReader(utils.ShipmentControlOfficialSealPNG))
	}
	if pdf.Err() {
		return fmt.Errorf("register certificate images: %v", pdf.Error())
	}
	return nil
}

func drawCertHeader(pdf *fpdf.Fpdf, data ShipmentControlCertificateData, y float64) float64 {
	const h = 30.0
	fill(pdf, certNavy)
	pdf.Rect(certSheetX0, y, certSheetX1-certSheetX0, h, "F")

	// Logo (left).
	if len(utils.LBOneTrackLogoPNG) > 0 {
		info := pdf.GetImageInfo("lb_logo")
		if info != nil && info.Height() > 0 {
			logoH := 11.0
			logoW := logoH * info.Width() / info.Height()
			if logoW > 55 {
				logoW = 55
				logoH = logoW * info.Height() / info.Width()
			}
			pdf.ImageOptions("lb_logo", certContentX0, y+(h-logoH)/2, logoW, logoH, false,
				fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		}
	}

	// Right block: certificate number, title, version line.
	pdf.SetXY(certContentX0+70, y+6)
	textColor(pdf, certAccent)
	setFont(pdf, "B", 8)
	pdf.CellFormat(certContentW-70, 4, tr(pdf, "CERTIFICADO N° "+data.ControlNumber), "", 2, "R", false, 0, "")

	textColor(pdf, rgb{255, 255, 255})
	setFont(pdf, "B", 15)
	pdf.CellFormat(certContentW-70, 7, tr(pdf, "Certificado de Control de Embarque"), "", 2, "R", false, 0, "")

	textColor(pdf, certLightText)
	setFont(pdf, "", 8)
	pdf.CellFormat(certContentW-70, 5, tr(pdf, "Versión 4 · Emitido el "+data.IssuedDate), "", 2, "R", false, 0, "")

	return y + h
}

func drawCompanies(pdf *fpdf.Fpdf, data ShipmentControlCertificateData, y float64) float64 {
	y += 6
	colW := certContentW / 2
	midX := certContentX0 + colW

	// Vertical separator between the two columns.
	draw(pdf, certBorder)
	pdf.SetLineWidth(0.2)

	leftY := drawCompanyColumn(pdf, certContentX0, y, colW-6, "Empresa Certificadora", [][2]string{
		{"Empresa", data.CertificadoraName},
		{"RUT", data.CertificadoraRUT},
		{"N° Registro SUBTEL", data.SubtelRegistry},
		{"N° de Control", data.ControlNumber},
	})
	rightY := drawCompanyColumn(pdf, midX+6, y, colW-6, "Empresa Solicitante", [][2]string{
		{"Empresa", data.SolicitanteName},
		{"RUT", data.SolicitanteRUT},
		{"Dirección", data.SolicitanteAddress},
	})

	bottom := leftY
	if rightY > bottom {
		bottom = rightY
	}
	pdf.Line(midX, y, midX, bottom-4)
	return bottom
}

func drawCompanyColumn(pdf *fpdf.Fpdf, x, y, w float64, heading string, fields [][2]string) float64 {
	y = drawSectionHeading(pdf, x, y, w, heading)
	for _, f := range fields {
		y = drawLabelValue(pdf, x, y, w, f[0], f[1])
		y += 3
	}
	return y
}

func drawDeviceData(pdf *fpdf.Fpdf, data ShipmentControlCertificateData, y float64) float64 {
	y = drawSectionHeading(pdf, certContentX0, y, certContentW, "Datos del Dispositivo")
	cols := [][2]string{
		{"Marca", data.Brand},
		{"Tipo", data.DeviceType},
		{"Versión OS", data.OsVersion},
		{"Versión HW", data.HardwareVersion},
		{"Versión SW", data.SoftwareVersion},
	}
	colW := certContentW / float64(len(cols))
	maxY := y
	for i, c := range cols {
		cx := certContentX0 + float64(i)*colW
		cy := drawLabelValue(pdf, cx, y, colW-3, c[0], c[1])
		if cy > maxY {
			maxY = cy
		}
	}
	return maxY
}

func drawGuides(pdf *fpdf.Fpdf, data ShipmentControlCertificateData, y float64) float64 {
	y = drawSectionHeading(pdf, certContentX0, y, certContentW, "Guías del Embarque")

	headers := []string{"Fecha", "Registro OABI", "Modelo", "Rework", "Cantidad registrada"}
	widths := []float64{0.18, 0.22, 0.24, 0.14, 0.22}
	values := []string{
		data.ShipmentDate,
		data.RegistroOABI,
		data.CommercialModel,
		data.ReworkNumber,
		data.RegisteredImeiCount + " IMEIs",
	}

	// Header row.
	fill(pdf, certTableHead)
	textColor(pdf, certNavy)
	setFont(pdf, "B", 8)
	x := certContentX0
	for i, hh := range headers {
		w := widths[i] * certContentW
		pdf.SetXY(x, y)
		pdf.CellFormat(w, 8, tr(pdf, hh), "", 0, "L", true, 0, "")
		x += w
	}
	y += 8

	// Value row with a top border.
	draw(pdf, certBorder)
	pdf.SetLineWidth(0.2)
	pdf.Line(certContentX0, y, certContentX1, y)
	textColor(pdf, certInk)
	setFont(pdf, "B", 9)
	x = certContentX0
	for i, v := range values {
		w := widths[i] * certContentW
		pdf.SetXY(x, y+1)
		pdf.CellFormat(w, 8, tr(pdf, v), "", 0, "L", false, 0, "")
		x += w
	}
	return y + 10
}

func drawChecklist(pdf *fpdf.Fpdf, y float64) float64 {
	y = drawSectionHeading(pdf, certContentX0, y, certContentW, "Checklist de Embarque")
	textColor(pdf, certMuted)
	setFont(pdf, "", 8)
	pdf.SetXY(certContentX0, y)
	pdf.CellFormat(certContentW, 4, tr(pdf, "Comparación Embarque vs Reporte de Homologación"), "", 2, "L", false, 0, "")
	y += 6

	left := []string{"Fabricante", "Marca", "Modelo"}
	right := []string{"Sistema Operativo", "Etiquetas", "Verificación IMEIs en Embarque"}
	rowH := 6.0
	for i := 0; i < 3; i++ {
		drawCheckItem(pdf, certContentX0, y, left[i])
		drawCheckItem(pdf, certContentX0+certContentW/2, y, right[i])
		y += rowH
	}
	return y
}

func drawCheckItem(pdf *fpdf.Fpdf, x, y float64, label string) {
	r := 2.2
	cx := x + r
	cy := y + 2.4
	fill(pdf, certAccent)
	pdf.Circle(cx, cy, r, "F")
	// Checkmark drawn as two strokes so it does not depend on a glyph.
	draw(pdf, certNavy)
	pdf.SetLineWidth(0.4)
	pdf.Line(cx-1.1, cy, cx-0.3, cy+0.9)
	pdf.Line(cx-0.3, cy+0.9, cx+1.2, cy-1.0)

	textColor(pdf, certBody)
	setFont(pdf, "", 9)
	pdf.SetXY(x+2*r+2, y)
	pdf.CellFormat(certContentW/2-2*r-4, 5, tr(pdf, label), "", 0, "L", false, 0, "")
}

func drawLegalAndSignature(pdf *fpdf.Fpdf, data ShipmentControlCertificateData, y float64) float64 {
	legal := "Asimismo, por medio de este certificado se acredita que el embarque antes descrito ha sido " +
		"sometido a las pruebas de control contempladas en el denominado Protocolo Básico de Homologación, " +
		"las cuales han resultado exitosas, y según lo cual se entenderá que los dispositivos que cumplan las " +
		"características contenidas en este certificado cumplen cabalmente con la Normativa de SUBTEL para poder " +
		"ser comercializados en el país."
	legal2 := "Este certificado es intransferible, se emiten 2 originales de este documento, en caso de requerir " +
		"copias adicionales por favor solicitarlos a contacto@lbtechnology-la.com indicando el código del certificado."

	textColor(pdf, certMuted)
	setFont(pdf, "", 8.5)
	pdf.SetXY(certContentX0, y)
	pdf.MultiCell(certContentW, 4.4, tr(pdf, legal), "", "L", false)
	pdf.Ln(1)
	pdf.SetX(certContentX0)
	pdf.MultiCell(certContentW, 4.4, tr(pdf, legal2), "", "L", false)

	sigY := pdf.GetY() + 6
	colW := certContentW / 2

	// Signature (left).
	textColor(pdf, certInk)
	setFont(pdf, "", 9)
	pdf.SetXY(certContentX0, sigY)
	pdf.CellFormat(colW, 5, tr(pdf, "Atentamente,"), "", 2, "L", false, 0, "")
	setFontStyle(pdf, "I", 20)
	pdf.SetXY(certContentX0, sigY+5)
	pdf.CellFormat(colW, 9, tr(pdf, "Glajay Tovar"), "", 2, "L", false, 0, "")
	draw(pdf, certInk)
	pdf.SetLineWidth(0.3)
	pdf.Line(certContentX0, sigY+16, certContentX0+45, sigY+16)
	textColor(pdf, certNavy)
	setFont(pdf, "B", 10)
	pdf.SetXY(certContentX0, sigY+17)
	pdf.CellFormat(colW, 5, tr(pdf, "Glajay Tovar"), "", 2, "L", false, 0, "")
	textColor(pdf, certMuted)
	setFont(pdf, "", 8.5)
	for _, line := range []string{"Project Manager Assistant", "LB Technology SPA", "Santiago, Chile"} {
		pdf.SetX(certContentX0)
		pdf.CellFormat(colW, 4.4, tr(pdf, line), "", 2, "L", false, 0, "")
	}

	// Official seal (right).
	if len(utils.ShipmentControlOfficialSealPNG) > 0 {
		info := pdf.GetImageInfo("official_seal")
		if info != nil && info.Height() > 0 {
			sealSize := 34.0
			sx := certContentX0 + colW + (colW-sealSize)/2
			pdf.ImageOptions("official_seal", sx, sigY+2, sealSize, sealSize, false,
				fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		}
	}
	return sigY + 40
}

func drawFooter(pdf *fpdf.Fpdf, data ShipmentControlCertificateData) {
	barH := 12.0
	noteH := 8.0
	barY := 285.0 - barH
	noteY := barY - noteH

	fill(pdf, rgb{255, 255, 255})
	pdf.Rect(certSheetX0, noteY, certSheetX1-certSheetX0, noteH, "F")
	textColor(pdf, rgb{148, 163, 184})
	setFont(pdf, "", 8)
	pdf.SetXY(certSheetX0, noteY+2)
	pdf.CellFormat(certSheetX1-certSheetX0, 4, tr(pdf, "LB Technology – Certificado Control de Embarque v4"), "", 0, "C", false, 0, "")

	fill(pdf, certNavy)
	pdf.Rect(certSheetX0, barY, certSheetX1-certSheetX0, barH, "F")
	textColor(pdf, rgb{148, 163, 184})
	setFont(pdf, "", 8)
	pdf.SetXY(certContentX0, barY+4)
	pdf.CellFormat(certContentW/2, 5, tr(pdf, fmt.Sprintf("© %d LB Technology. Todos los derechos reservados.", data.Year)), "", 0, "L", false, 0, "")
	textColor(pdf, certAccent)
	setFont(pdf, "B", 9)
	pdf.SetXY(certContentX0+certContentW/2, barY+4)
	pdf.CellFormat(certContentW/2, 5, tr(pdf, data.ControlNumber), "", 0, "R", false, 0, "")
}

// ---- small drawing helpers ----

func drawSectionHeading(pdf *fpdf.Fpdf, x, y, w float64, text string) float64 {
	textColor(pdf, certNavy)
	setFont(pdf, "B", 9)
	pdf.SetXY(x, y)
	pdf.CellFormat(w, 5, tr(pdf, text), "", 2, "L", false, 0, "")
	return y + 7
}

func drawLabelValue(pdf *fpdf.Fpdf, x, y, w float64, label, value string) float64 {
	if value == "" {
		value = "—"
	}
	textColor(pdf, certMuted)
	setFont(pdf, "B", 7.5)
	pdf.SetXY(x, y)
	pdf.CellFormat(w, 4, tr(pdf, label), "", 2, "L", false, 0, "")
	textColor(pdf, certInk)
	setFont(pdf, "B", 9.5)
	pdf.SetX(x)
	pdf.MultiCell(w, 4.6, tr(pdf, value), "", "L", false)
	return pdf.GetY()
}

func drawAccentDivider(pdf *fpdf.Fpdf, y float64) float64 {
	fill(pdf, certAccent)
	pdf.Rect(certSheetX0, y, certSheetX1-certSheetX0, 0.8, "F")
	return y + 0.8
}

func drawInnerDivider(pdf *fpdf.Fpdf, y float64) float64 {
	fill(pdf, certAccent)
	pdf.Rect(certContentX0, y, certContentW, 0.6, "F")
	return y + 0.6
}

func fill(pdf *fpdf.Fpdf, c rgb)      { pdf.SetFillColor(c.r, c.g, c.b) }
func draw(pdf *fpdf.Fpdf, c rgb)      { pdf.SetDrawColor(c.r, c.g, c.b) }
func textColor(pdf *fpdf.Fpdf, c rgb) { pdf.SetTextColor(c.r, c.g, c.b) }

func setFont(pdf *fpdf.Fpdf, style string, size float64) {
	pdf.SetFont("Helvetica", style, size)
}

func setFontStyle(pdf *fpdf.Fpdf, style string, size float64) {
	pdf.SetFont("Helvetica", style, size)
}

// tr converts UTF-8 text to the encoding fpdf core fonts expect (cp1252),
// so accented Spanish characters render correctly.
func tr(pdf *fpdf.Fpdf, s string) string {
	return pdf.UnicodeTranslatorFromDescriptor("")(s)
}
