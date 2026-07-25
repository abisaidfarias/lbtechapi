package functions

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// controlNumberPattern matches a well-formed control number: an alphanumeric
// client code, a dash, then 9 digits (YYMMDD + 3-digit sequence).
var controlNumberPattern = regexp.MustCompile(`^[A-Za-z0-9]+-[0-9]{9}$`)

// IsValidControlNumber reports whether s has the expected control-number shape.
// It rejects values corrupted by URL-encoding (e.g. a stray "%25") so callers
// can rebuild a clean number from the company client id instead of reusing junk.
func IsValidControlNumber(s string) bool {
	return controlNumberPattern.MatchString(strings.TrimSpace(s))
}

const (
	ShipmentCertificadoraName           = "LB Technology SPA"
	ShipmentCertificadoraRUT            = "76.527.872-4"
	ShipmentCertificadoraSubtelRegistry = "110"
)

// ShipmentControlCertificateData is passed to the certificate HTML template.
type ShipmentControlCertificateData struct {
	LBLogoDataURI       template.URL
	OfficialSealDataURI template.URL
	CompanyLogoURL      string
	ChileGovLogoURL    string
	ControlNumber      string
	CertificadoraName  string
	CertificadoraRUT   string
	SubtelRegistry     string
	SolicitanteName    string
	SolicitanteRUT     string
	SolicitanteAddress string
	Brand              string
	DeviceType         string
	OsVersion          string
	HardwareVersion    string
	SoftwareVersion    string
	ShipmentDate       string
	IssuedDate         string
	RegistroOABI       string
	CommercialModel    string
	ReworkNumber       string
	RegisteredImeiCount string
	Year               int
}

func BuildShipmentControlCertificateData(
	company *responses.Company,
	multibanda *responses.MultibandaExpanded,
	controlNumber string,
	registroOABI string,
	registeredImeiCount int,
	reworkNumber string,
) ShipmentControlCertificateData {
	solicitanteName := strings.TrimSpace(company.RazonSocial)
	if solicitanteName == "" {
		solicitanteName = strings.TrimSpace(company.Name)
	}

	now := time.Now()

	deviceType := strings.TrimSpace(multibanda.Device.Type)
	if deviceType == "" {
		deviceType = "—"
	}

	return ShipmentControlCertificateData{
		LBLogoDataURI:       shipmentCertificateLogoDataURI(),
		OfficialSealDataURI: shipmentCertificateOfficialSealDataURI(),
		CompanyLogoURL:      strings.TrimSpace(company.LogoUrl),
		ChileGovLogoURL:    strings.TrimSpace(os.Getenv("SHIPMENT_CERT_CHILE_LOGO_URL")),
		ControlNumber:      controlNumber,
		CertificadoraName:  ShipmentCertificadoraName,
		CertificadoraRUT:   ShipmentCertificadoraRUT,
		SubtelRegistry:     ShipmentCertificadoraSubtelRegistry,
		SolicitanteName:    solicitanteName,
		SolicitanteRUT:     strings.TrimSpace(company.Rut),
		SolicitanteAddress: strings.TrimSpace(company.Address),
		Brand:              strings.TrimSpace(multibanda.Brand.Name),
		DeviceType:         deviceType,
		OsVersion: FormatMultibandaOsVersion(
			multibanda.OsVersionView,
			multibanda.Device.PlatformOs,
			multibanda.OsVersion,
		),
		HardwareVersion: strings.TrimSpace(multibanda.HardwareVersion),
		SoftwareVersion: strings.TrimSpace(multibanda.SoftwareVersion),
		ShipmentDate:    formatShipmentCertificateDate(now),
		IssuedDate:      formatShipmentCertificateIssuedDate(now),
		RegistroOABI:    strings.TrimSpace(registroOABI),
		CommercialModel: strings.TrimSpace(multibanda.Device.CommercialModel),
		ReworkNumber:    strings.TrimSpace(reworkNumber),
		RegisteredImeiCount: formatShipmentImeiQuantity(registeredImeiCount),
		Year:            now.Year(),
	}
}

func RenderShipmentControlCertificateHTML(data ShipmentControlCertificateData, templatePath string) ([]byte, error) {
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, err
	}
	var htmlBuf bytes.Buffer
	if err := t.Execute(&htmlBuf, data); err != nil {
		return nil, err
	}
	html := htmlBuf.Bytes()
	html = bytes.ReplaceAll(html, []byte("\r\n"), []byte("\n"))
	html = bytes.ReplaceAll(html, []byte("\n"), []byte("\r\n"))
	return html, nil
}

func BuildShipmentControlControlNumber(clientID string, sequence int, now time.Time) string {
	clientID = strings.TrimSpace(clientID)
	datePart := now.In(shipmentCertificateLocation()).Format("060102")
	return fmt.Sprintf("%s-%s%03d", clientID, datePart, sequence)
}

func ShipmentControlControlNumberPrefix(clientID string, now time.Time) string {
	clientID = strings.TrimSpace(clientID)
	datePart := now.In(shipmentCertificateLocation()).Format("060102")
	return fmt.Sprintf("%s-%s", clientID, datePart)
}

func shipmentCertificateLogoDataURI() template.URL {
	publicLogoURL := strings.TrimSpace(os.Getenv("TRACKING_LOGO_URL"))
	if publicLogoURL != "" {
		return template.URL(publicLogoURL)
	}
	if len(utils.LBOneTrackLogoPNG) == 0 {
		return ""
	}
	return template.URL("data:image/png;base64," + base64.StdEncoding.EncodeToString(utils.LBOneTrackLogoPNG))
}

func shipmentCertificateOfficialSealDataURI() template.URL {
	if customURL := strings.TrimSpace(os.Getenv("SHIPMENT_CERT_OFFICIAL_SEAL_URL")); customURL != "" {
		return template.URL(customURL)
	}
	if len(utils.ShipmentControlOfficialSealPNG) == 0 {
		return ""
	}
	return template.URL("data:image/png;base64," + base64.StdEncoding.EncodeToString(utils.ShipmentControlOfficialSealPNG))
}

func formatShipmentCertificateDate(value time.Time) string {
	return value.In(shipmentCertificateLocation()).Format("2006-01-02")
}

func formatShipmentCertificateIssuedDate(value time.Time) string {
	return value.In(shipmentCertificateLocation()).Format("02-01-2006")
}

func formatShipmentImeiQuantity(value int) string {
	if value <= 0 {
		return "0"
	}
	s := strconv.Itoa(value)
	if len(s) <= 3 {
		return s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	if s != "" {
		parts = append([]string{s}, parts...)
	}
	return strings.Join(parts, ".")
}

func shipmentCertificateLocation() *time.Location {
	loc, err := time.LoadLocation("America/Santiago")
	if err != nil {
		return time.FixedZone("CLT", -3*60*60)
	}
	return loc
}

// ControlNumberFromCertificateURL extracts the control number from the S3 object key
// (e.g. .../002-260717001.pdf → 002-260717001).
func ControlNumberFromCertificateURL(certificateURL string) string {
	certificateURL = strings.TrimSpace(certificateURL)
	if certificateURL == "" {
		return ""
	}
	certificateURL = strings.Split(certificateURL, "?")[0]
	base := certificateURL[strings.LastIndex(certificateURL, "/")+1:]
	base = strings.TrimSuffix(base, ".pdf")
	// S3 URLs percent-encode the key, so decode before returning to avoid a
	// stray "%25" leaking back into the control number.
	if decoded, err := url.QueryUnescape(base); err == nil {
		base = decoded
	}
	return base
}

// CacheBustCertificateURL appends a version query param so browsers fetch the latest PDF
// after the same S3 object key is overwritten on regeneration.
func CacheBustCertificateURL(certificateURL string) string {
	certificateURL = strings.TrimSpace(certificateURL)
	if certificateURL == "" {
		return ""
	}
	base := strings.Split(certificateURL, "?")[0]
	return fmt.Sprintf("%s?v=%d", base, time.Now().Unix())
}
