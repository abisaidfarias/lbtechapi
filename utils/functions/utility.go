package functions

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"mime"
	"mime/multipart"
	"net"
	"net/smtp"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/config"
	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	utils "github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofrs/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

func UpdateCodeVersion(code string) string {
	size := len(code)
	char := []rune(code)[size-1]
	nextVersion := NextRune(char)
	res := code[:size-1] + string(nextVersion)
	return res
}

func NextRune(r rune) rune {
	return r + 1
}
func HashPassword(password string) string {

	passwordBytes := []byte(password)

	hash, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)

	if err != nil {
		panic(err)
	}

	return string(hash)
}

func StringsToObjectIds(ids []string) []primitive.ObjectID {
	var objectIds []primitive.ObjectID
	for _, id := range ids {
		objID, _ := primitive.ObjectIDFromHex(id)
		objectIds = append(objectIds, objID)
	}
	return objectIds
}
func GetAccountInfo() (string, string, string, string) {
	azrKey := "CcvX9B5daFIR1ZmUbSqyUWU5LG0GChJ5BfElplqTwR3ZXubbEPgqOWrtRccutzkJ6hEtEsIWPsndJJKzIxfQVA=="
	azrBlobAccountName := "lbtechfilestorage"
	azrPrimaryBlobServiceEndpoint := fmt.Sprintf("https://%s.blob.core.windows.net/", azrBlobAccountName)
	azrBlobContainer := "blog-photos"

	return azrKey, azrBlobAccountName, azrPrimaryBlobServiceEndpoint, azrBlobContainer
}
func GetBlobName() string {
	t := time.Now()
	uuid, _ := uuid.NewV4()

	return fmt.Sprintf("%s-%v.jpg", t.Format("20060102"), uuid)
}
func RandomFileName(mimeType string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%s%s", strconv.Itoa(r.Int()), mimeType)
}

// test cases

func GenerateCategoryMap(categories []*responses.TestCategory) map[string]string {
	m := make(map[string]string)

	for _, v := range categories {
		m[v.Name] = v.ID.Hex()
	}

	return m
}

func GenerateTestCaseFromLine(line []string, catId string) *models.TestCase {

	var testCase models.TestCase = models.TestCase{}

	objID, err := primitive.ObjectIDFromHex(catId)

	if err != nil {
		return nil
	}

	testCase.Code = line[0]
	testCase.Description = line[1]
	testCase.Expected = line[2]
	testCase.Name = line[3]
	testCase.TestCategory = objID
	testCase.IsActive = true

	return &testCase
}

func SetHomologationDatesToNull(homologation *responses.HomologationExpanded) {
	if homologation.PlanningDate.IsZero() {
		homologation.PlanningDate = nil
	}
	if homologation.SampleStartDate.IsZero() {
		homologation.SampleStartDate = nil
	}
	if homologation.SampleEndDate.IsZero() {
		homologation.SampleEndDate = nil
	}
	if homologation.TestStartDate.IsZero() {
		homologation.TestStartDate = nil
	}
	if homologation.TestEndDate.IsZero() {
		homologation.TestEndDate = nil
	}
	if homologation.UnderStartDate.IsZero() {
		homologation.UnderStartDate = nil
	}
	if homologation.UnderEndDate.IsZero() {
		homologation.UnderEndDate = nil
	}
	if homologation.CompletedDate.IsZero() {
		homologation.CompletedDate = nil
	}
}
func ValidateUserCredentials(passwordHash string, password string) error {

	passwordMatches := CompareHashAndPassword(passwordHash, []byte(password))

	if !passwordMatches {
		return fmt.Errorf("%w", utils.ErrorInvalidCredentials)
	}

	return nil
}
func GenerateJWT(user *responses.AuthUser) string {

	var JWTKey = []byte(config.GetValue("SECRET_KEY"))
	expirationTime := time.Now().Add(1000 * time.Hour)

	claims := &models.AuthClaim{
		ID:         user.ID.Hex(),
		CompanyID:  user.Company.Hex(),
		IsInternal: strconv.FormatBool(user.IsInternal),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(JWTKey)

	if err != nil {
		panic(err)
	}

	return tokenString
}
func CompareHashAndPassword(hashedPassword string, incomingPassword []byte) bool {

	byteHash := []byte(hashedPassword)
	err := bcrypt.CompareHashAndPassword(byteHash, incomingPassword)
	return err == nil
}
func SendNotifications(toList []string, body bytes.Buffer) {

	from := config.GetValue("EMAIL_FROM")
	password := config.GetValue("EMAIL_PASSWORD")

	smtpHost := config.GetValue("SMTP_CLIENTE")
	smtpPort := config.GetValue("EMAIL_PORT")

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", smtpHost, smtpPort))
	if err != nil {
		return
	}

	c, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		println(err)
	}

	tlsconfig := &tls.Config{
		ServerName: smtpHost,
	}

	if err = c.StartTLS(tlsconfig); err != nil {
		println(err)
	}

	auth := LoginAuth(from, password)

	if err = c.Auth(auth); err != nil {
		println(err)
	}
	// Sending email.
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, toList, body.Bytes())
	if err != nil {
		fmt.Println(err)
		return
	}
}

const trackingMoveLogoCID = "lbonetrack-logo"

// buildMoveEmailMultipartRelated builds multipart/related (HTML + inline PNG) using mime/multipart.
func buildMoveEmailMultipartRelated(html []byte, logoPNG []byte) ([]byte, error) {
	var mpBody bytes.Buffer
	w := multipart.NewWriter(&mpBody)
	boundary := w.Boundary()

	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", "text/html; charset=UTF-8")
	h.Set("Content-Transfer-Encoding", "8bit")
	htmlPart, err := w.CreatePart(h)
	if err != nil {
		return nil, err
	}
	if _, err := htmlPart.Write(html); err != nil {
		return nil, err
	}

	h2 := make(textproto.MIMEHeader)
	h2.Set("Content-Type", "image/png")
	h2.Set("Content-Transfer-Encoding", "base64")
	h2.Set("Content-Disposition", `inline; filename="lbonetrack_logo.png"`)
	h2.Set("Content-Id", "<"+trackingMoveLogoCID+">")
	imgPart, err := w.CreatePart(h2)
	if err != nil {
		return nil, err
	}
	enc := base64.NewEncoder(base64.StdEncoding, imgPart)
	if _, err := enc.Write(logoPNG); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	var msg bytes.Buffer
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: multipart/related; type=\"text/html\"; boundary=\"" + boundary + "\"\r\n\r\n")
	msg.Write(mpBody.Bytes())
	return msg.Bytes(), nil
}

// writeMoveEmailMainPart writes a multipart/mixed part: either text/html or multipart/related (HTML + inline logo).
func writeMoveEmailMainPart(mw *multipart.Writer, html []byte, logoPNG []byte, publicLogoURL string) error {
	if publicLogoURL != "" || len(logoPNG) == 0 {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Type", "text/html; charset=UTF-8")
		h.Set("Content-Transfer-Encoding", "8bit")
		p, err := mw.CreatePart(h)
		if err != nil {
			return err
		}
		_, err = p.Write(html)
		return err
	}
	var nested bytes.Buffer
	nw := multipart.NewWriter(&nested)
	h1 := make(textproto.MIMEHeader)
	h1.Set("Content-Type", "text/html; charset=UTF-8")
	h1.Set("Content-Transfer-Encoding", "8bit")
	w1, err := nw.CreatePart(h1)
	if err != nil {
		return err
	}
	if _, err := w1.Write(html); err != nil {
		return err
	}
	h2 := make(textproto.MIMEHeader)
	h2.Set("Content-Type", "image/png")
	h2.Set("Content-Transfer-Encoding", "base64")
	h2.Set("Content-Disposition", `inline; filename="lbonetrack_logo.png"`)
	h2.Set("Content-Id", "<"+trackingMoveLogoCID+">")
	w2, err := nw.CreatePart(h2)
	if err != nil {
		return err
	}
	enc := base64.NewEncoder(base64.StdEncoding, w2)
	if _, err := enc.Write(logoPNG); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}
	if err := nw.Close(); err != nil {
		return err
	}
	hOuter := make(textproto.MIMEHeader)
	hOuter.Set("Content-Type", fmt.Sprintf(`multipart/related; type="text/html"; boundary="%s"`, nw.Boundary()))
	relPart, err := mw.CreatePart(hOuter)
	if err != nil {
		return err
	}
	_, err = relPart.Write(nested.Bytes())
	return err
}

// RenderTrackingMoveEmailHTML renders the movement email template to HTML bytes (caller sets LogoDataURI).
func RenderTrackingMoveEmailHTML(data TrackingMoveEmailData, templatePath string) ([]byte, error) {
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

// SendNotificationsMoveEmail sends the movement HTML.
// If pdfAttachment is non-empty, the message is multipart/mixed: HTML body (same rules as below) + PDF attachment.
// Logo resolution order:
//  1. TRACKING_LOGO_URL (https URL to a public PNG) — most reliable in Gmail when images are allowed.
//  2. Else embedded PNG from logoPNG as multipart/related + cid: (built with mime/multipart).
//  3. Else HTML only (no logo).
func SendNotificationsMoveEmail(toList []string, subject string, data TrackingMoveEmailData, templatePath string, logoPNG []byte, pdfAttachment []byte, pdfFileName string) error {
	from := config.GetValue("EMAIL_FROM")
	password := config.GetValue("EMAIL_PASSWORD")
	smtpHost := config.GetValue("SMTP_CLIENTE")
	smtpPort := config.GetValue("EMAIL_PORT")

	publicLogoURL := strings.TrimSpace(os.Getenv("TRACKING_LOGO_URL"))

	d := data
	switch {
	case publicLogoURL != "":
		d.LogoDataURI = template.URL(publicLogoURL)
	case len(logoPNG) > 0:
		d.LogoDataURI = template.URL("cid:" + trackingMoveLogoCID)
	default:
		d.LogoDataURI = ""
	}

	html, err := RenderTrackingMoveEmailHTML(d, templatePath)
	if err != nil {
		return err
	}

	subjectValue := strings.TrimSpace(strings.TrimPrefix(subject, "Subject:"))
	subjectValue = strings.TrimSpace(subjectValue)

	var msg bytes.Buffer
	msg.WriteString("From: " + from + "\r\n")
	msg.WriteString("To: " + strings.Join(toList, ", ") + "\r\n")
	msg.WriteString("Subject: " + subjectValue + "\r\n")

	if len(pdfAttachment) == 0 {
		var mimeBody bytes.Buffer
		if publicLogoURL != "" || len(logoPNG) == 0 {
			mimeBody.WriteString("MIME-Version: 1.0\r\n")
			mimeBody.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
			mimeBody.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
			mimeBody.Write(html)
		} else {
			part, err := buildMoveEmailMultipartRelated(html, logoPNG)
			if err != nil {
				return err
			}
			if _, err := mimeBody.Write(part); err != nil {
				return err
			}
		}
		if _, err := io.Copy(&msg, &mimeBody); err != nil {
			return err
		}
	} else {
		var mixedBody bytes.Buffer
		mw := multipart.NewWriter(&mixedBody)
		boundary := mw.Boundary()
		if err := writeMoveEmailMainPart(mw, html, logoPNG, publicLogoURL); err != nil {
			return err
		}
		fn := strings.TrimSpace(pdfFileName)
		if fn == "" {
			fn = "move-report.pdf"
		}
		pdfh := make(textproto.MIMEHeader)
		pdfh.Set("Content-Type", "application/pdf")
		pdfh.Set("Content-Transfer-Encoding", "base64")
		pdfh.Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": fn}))
		pw, err := mw.CreatePart(pdfh)
		if err != nil {
			return err
		}
		b64 := base64.NewEncoder(base64.StdEncoding, pw)
		if _, err := b64.Write(pdfAttachment); err != nil {
			return err
		}
		if err := b64.Close(); err != nil {
			return err
		}
		if err := mw.Close(); err != nil {
			return err
		}
		msg.WriteString("MIME-Version: 1.0\r\n")
		msg.WriteString("Content-Type: multipart/mixed; boundary=" + boundary + "\r\n\r\n")
		msg.Write(mixedBody.Bytes())
	}

	auth := LoginAuth(from, password)
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, toList, msg.Bytes())
}

func GetEmails(isInternal bool, companyId primitive.ObjectID) ([]string, bool) {

	var toList []string
	if !isInternal {
		notificationCollection := database.GetInstance().Collection("notifications")

		var notification *responses.Notification
		err := notificationCollection.FindOne(context.TODO(),
			queries.GetNotifictionByCompany(companyId)).Decode(&notification)
		if err != nil {
			return nil, true
		}
		for _, user := range notification.NotificationEmails {
			toList = append(toList, user.Email)
		}
	} else {
		userCollection := database.GetInstance().Collection("users")

		var user *responses.UserExpanded

		cursor, err := userCollection.Aggregate(context.TODO(), queries.GetUserExpandedInternal())
		if err != nil {
			return nil, true
		}
		for cursor.Next(context.TODO()) {
			cursor.Decode(&user)
			toList = append(toList, user.Email)
		}
	}
	if len(toList) == 0 {
		return nil, true
	}
	return toList, false
}
func GetHomologationBodyMessage(subject string, mainMessge string, projectType string, brand string, technicalModel string,
	commercialModel string, softwareVersion string, osVersion string, country string,
	homologationType string, carrier string, testingType string, planningDate string,
	sampleStartDate string, sampleEndDate string, testStartDate string,
	testEndDate string, underStartDate string, underEndDate string,
	resultDate string, templatePath string, userName string,
	finished bool, desicion string) (bytes.Buffer, error) {

	var body bytes.Buffer
	body.Write([]byte(fmt.Sprintf("%s \n%s\n\n", subject, utils.MIME_HEADERS)))

	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return body, err
	}
	if err != nil {
		return body, err
	}
	t.Execute(&body, struct {
		MainMessage     string
		Date            string
		ProjectType     string
		Brand           string
		TechnicalModel  string
		CommercialModel string
		SoftwareVersion string
		OSversion       string
		Country         string
		ApprovalType    string
		Carrier         string
		TestingType     string
		PlanningDate    string
		SampleStart     string
		SampleEndDate   string
		TestStart       string
		TestEndDate     string
		UnderStart      string
		UnderEndDate    string
		ResultDate      string
		UserName        string
		Finished        bool
		Decision        string
	}{
		MainMessage:     mainMessge,
		Date:            fmt.Sprintf("%02d/%02d/%d", time.Now().Day(), time.Now().Month(), time.Now().Year()),
		ProjectType:     projectType,
		Brand:           brand,
		TechnicalModel:  technicalModel,
		CommercialModel: commercialModel,
		SoftwareVersion: softwareVersion,
		OSversion:       osVersion,
		Country:         country,
		ApprovalType:    homologationType,
		Carrier:         carrier,
		TestingType:     testingType,
		PlanningDate:    planningDate,
		SampleStart:     sampleStartDate,
		SampleEndDate:   sampleEndDate,
		TestStart:       testStartDate,
		TestEndDate:     testEndDate,
		UnderStart:      underStartDate,
		UnderEndDate:    underEndDate,
		ResultDate:      resultDate,
		UserName:        userName,
		Finished:        finished,
		Decision:        desicion,
	})

	return body, nil
}

// TrackingEmailDeviceRow is one horizontal row in the tracking email table (one IMEI per row).
type TrackingEmailDeviceRow struct {
	Client              string
	Country             string
	Brand               string
	TechnicalModel      string
	CommercialModel     string
	Imei                string
	PreviousLocation    string
	NewLocation         string
	LBResponsible       string
	ExternalResponsible string
	Comments            string
	RegistrationDate    string
	RegisteredBy        string
}

// TrackingMoveEmailData is passed to the movement notification HTML template.
type TrackingMoveEmailData struct {
	// LogoDataURI is set at send time to cid:lbonetrack-logo (inline PNG) or empty when no logo.
	LogoDataURI         template.URL
	ClientName          string
	Country             string
	TrackingID          string
	TotalSamples        int
	PreviousLocation    string
	NewLocation         string
	RegistrationDate    string
	LBResponsible       string
	ExternalResponsible string
	RegisteredBy        string
	Comments            string
	Year                int
	MainMessage         string
	Rows                []TrackingEmailDeviceRow
	// In-store delivery confirmation (confirm-move-delivery); when true, template shows receiver + signature.
	HasDeliveryConfirmation bool
	ReceiverFullName        string
	ReceiverRUT             string
	ReceiverDeliveredAt     string
	ReceiverSignatureDataURI template.URL
}

func GetTrackingBodyMessage(subject string, mainMessge string, rows []TrackingEmailDeviceRow,
	templatePath string) (bytes.Buffer, error) {

	var body bytes.Buffer
	body.Write([]byte(fmt.Sprintf("%s \n%s\n\n", subject, utils.MIME_HEADERS)))

	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return body, err
	}
	t.Execute(&body, struct {
		MainMessage string
		Rows        []TrackingEmailDeviceRow
	}{
		MainMessage: mainMessge,
		Rows:        rows,
	})

	return body, nil
}

type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unknown From Server")
		}
	}
	return nil, nil
}
func GetNotificationMessageAndSubject(homologation *request.Homologation,
	country string, brand string, commercialModel string) (string, string) {
	var mainMessage string
	var subject string
	switch homologation.CurrentPhase {

	case 0:
		subject = fmt.Sprintf("Subject: Planning %s %s %s",
			country, brand, commercialModel)
		return utils.PLANNING_MAIN_MESSAGE, subject
	case 1:
		if homologation.SampleEndDate.IsZero() {
			subject = fmt.Sprintf("Subject: Sample Reception Start Date %s %s %s",
				country, brand, commercialModel)
			return utils.SAMPLE_START_MAIN_MESSAGE, subject
		}
		subject = fmt.Sprintf("Subject: Sample Reception End Date %s %s %s",
			country, brand, commercialModel)
		return utils.SAMPLE_END_MAIN_MESSAGE, subject
	case 2:
		if homologation.TestStartDate.IsZero() {
			subject = fmt.Sprintf("Subject: Sample Reception End Date %s %s %s",
				country, brand, commercialModel)
			return utils.SAMPLE_END_MAIN_MESSAGE, subject
		}
		subject = fmt.Sprintf("Subject: Test Start Date %s %s %s",
			country, brand, commercialModel)
		return utils.TEST_MAIN_MESSAGE, subject
	case 3:
		subject = fmt.Sprintf("Subject: Test End Date %s %s %s",
			country, brand, commercialModel)
		return utils.UNDER_MAIN_MESSAGE, subject
	case 4:

		if enums.HomologationStatus_type[homologation.Status] == "Approved" {
			subject = fmt.Sprintf("Subject: Carrier Decision Approved %s %s %s",
				country, brand, commercialModel)
			return utils.APPROVED_MAIN_MESSAGE, subject
		}
		if enums.HomologationStatus_type[homologation.Status] == "Rejected" {
			subject = fmt.Sprintf("Subject: Carrier Decision Rejected %s %s %s",
				country, brand, commercialModel)
			return utils.REJECTED_MAIN_MESSAGE, subject
		}
		if enums.HomologationStatus_type[homologation.Status] == "Finished" {
			subject = fmt.Sprintf("Subject: Homologation Process Finished %s %s %s",
				country, brand, commercialModel)
			return utils.FINISHED_MAIN_MESSAGE, subject
		}
	}
	return mainMessage, subject
}
func GetFailBodyMessage(subject string, mainMessge string, client string, country string,
	brand string, technicalModel string, commercialModel string,
	softwareVersion string, osVersion string, homologationType string,
	projectType string, testCode string, testName string,
	issueOverview string, actualResult string,
	stepToReproduce string, expectedResult string, issueFrecuency int,
	issueSeverity int, hiperLink []request.Hyperlink,
	templatePath string, userName string) (bytes.Buffer, error) {

	var body bytes.Buffer
	body.Write([]byte(fmt.Sprintf("%s \n%s\n\n", subject, utils.MIME_HEADERS)))

	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return body, err
	}
	if err != nil {
		return body, err
	}
	t.Execute(&body, struct {
		MainMessage     string
		Date            string
		Client          string
		Country         string
		Brand           string
		TechnicalModel  string
		CommercialModel string
		SoftwareVersion string
		OSversion       string
		ApprovalType    string
		ProjectType     string
		TestCode        string
		TestName        string
		IssueOverview   string
		ActualResult    string
		StepToReproduce string
		ExpectedResult  string
		IssueFrecuency  string
		IssueSeverity   string
		Hyperlinks      []request.Hyperlink
		LinkDescription string
		UserName        string
	}{
		MainMessage:     mainMessge,
		Date:            fmt.Sprintf("%02d/%02d/%d", time.Now().Day(), time.Now().Month(), time.Now().Year()),
		ProjectType:     projectType,
		Brand:           brand,
		TechnicalModel:  technicalModel,
		CommercialModel: commercialModel,
		SoftwareVersion: softwareVersion,
		OSversion:       osVersion,
		Country:         country,
		ApprovalType:    homologationType,
		Client:          client,
		TestCode:        testCode,
		TestName:        testName,
		IssueOverview:   issueOverview,
		ActualResult:    actualResult,
		StepToReproduce: stepToReproduce,
		ExpectedResult:  expectedResult,
		IssueFrecuency:  enums.TestFailureFrequency_key[issueFrecuency],
		IssueSeverity:   enums.TestFailureSeverity_key[issueSeverity],
		Hyperlinks:      hiperLink,
		UserName:        userName,
	})

	return body, nil
}
func GetDeviceBodyMessage(subject string, mainMessge string, brand string,
	deviceRequest request.Device, templatePath string, userName string) (bytes.Buffer, error) {

	var body bytes.Buffer
	body.Write([]byte(fmt.Sprintf("%s \n%s\n\n", subject, utils.MIME_HEADERS)))

	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return body, err
	}
	if err != nil {
		return body, err
	}
	t.Execute(&body, struct {
		MainMessage            string
		Date                   string
		Type                   string
		Brand                  string
		CommercialModel        string
		TechnicalModel         string
		DisplayType            string
		DisplaySize            string
		DisplayResolution      string
		PlatformOs             string
		PlatformChipsetBrand   string
		PlatformChipsetModel   string
		PlatformCpu            string
		MemoryRom              string
		MemoryRam              string
		MemoryExtended         string
		MemoryCpu              string
		MemoryType             string
		CameraFront            string
		CameraBack             string
		CommunicationWlan      string
		CommunicationGps       string
		CommunicationNfc       string
		CommunicationRadio     string
		CommunicationUsb       string
		CommunicationBluetooth string
		BatteryType            string
		BatteryCapacity        string
		BatteryState           string
		BatteryInductedCharger string
		SensorFingerprint      string
		SensorProximity        string
		SensorAmbientLight     string
		SensorAccelerometer    string
		SensorGyroscope        string
		SensorHall             string
		NetworkGsm             string
		NetworkWcdma           string
		NetworkLte             string
		NetworkVolte           string
		NetworkVowifi          string
		NetworkVilte           string
		Network5g              string
		NetworkCarrierAgg      string
		ImageUrl               string
		SoftwareCode           string
		HardwareCode           string
		IngCode                string
		LoggingCode            string
		SimSupported           string
		SimType                string
		Esim                   string
		UserName               string
	}{
		MainMessage:            mainMessge,
		Date:                   fmt.Sprintf("%02d/%02d/%d", time.Now().Day(), time.Now().Month(), time.Now().Year()),
		Type:                   deviceRequest.Type,
		Brand:                  brand,
		CommercialModel:        deviceRequest.CommercialModel,
		TechnicalModel:         deviceRequest.TechnicalModel,
		DisplayType:            deviceRequest.DisplayType,
		DisplaySize:            deviceRequest.DisplaySize,
		DisplayResolution:      deviceRequest.DisplayResolution,
		PlatformOs:             deviceRequest.PlatformOs,
		PlatformChipsetBrand:   deviceRequest.PlatformChipsetBrand,
		PlatformChipsetModel:   deviceRequest.PlatformChipsetModel,
		PlatformCpu:            deviceRequest.PlatformCpu,
		MemoryRom:              deviceRequest.MemoryRom,
		MemoryRam:              deviceRequest.MemoryRam,
		MemoryExtended:         strconv.FormatBool(deviceRequest.MemoryExtended),
		MemoryCpu:              deviceRequest.MemoryCpu,
		MemoryType:             deviceRequest.MemoryType,
		CameraFront:            deviceRequest.CameraFront,
		CameraBack:             deviceRequest.CameraBack,
		CommunicationWlan:      strconv.FormatBool(deviceRequest.CommunicationWlan),
		CommunicationGps:       strconv.FormatBool(deviceRequest.CommunicationGps),
		CommunicationNfc:       strconv.FormatBool(deviceRequest.CommunicationNfc),
		CommunicationRadio:     strconv.FormatBool(deviceRequest.CommunicationRadio),
		CommunicationUsb:       deviceRequest.CommunicationUsb,
		CommunicationBluetooth: strconv.FormatBool(deviceRequest.CommunicationBluetooth),
		BatteryType:            deviceRequest.BatteryType,
		BatteryCapacity:        deviceRequest.BatteryCapacity,
		BatteryState:           deviceRequest.BatteryState,
		BatteryInductedCharger: strconv.FormatBool(deviceRequest.BatteryInductedCharger),
		SensorFingerprint:      strconv.FormatBool(deviceRequest.SensorFingerprint),
		SensorProximity:        strconv.FormatBool(deviceRequest.SensorProximity),
		SensorAmbientLight:     strconv.FormatBool(deviceRequest.SensorAmbientLight),
		SensorAccelerometer:    strconv.FormatBool(deviceRequest.SensorAccelerometer),
		SensorGyroscope:        strconv.FormatBool(deviceRequest.SensorGyroscope),
		SensorHall:             strconv.FormatBool(deviceRequest.SensorHall),
		NetworkGsm:             strconv.FormatBool(deviceRequest.NetworkGsm),
		NetworkWcdma:           strconv.FormatBool(deviceRequest.NetworkWcdma),
		NetworkLte:             strconv.FormatBool(deviceRequest.NetworkLte),
		NetworkVolte:           strconv.FormatBool(deviceRequest.NetworkVolte),
		NetworkVowifi:          strconv.FormatBool(deviceRequest.NetworkVowifi),
		NetworkVilte:           strconv.FormatBool(deviceRequest.NetworkVilte),
		Network5g:              strconv.FormatBool(deviceRequest.Network5g),
		NetworkCarrierAgg:      strconv.FormatBool(deviceRequest.NetworkCarrierAgg),
		ImageUrl:               deviceRequest.ImageUrl,
		SoftwareCode:           deviceRequest.SoftwareCode,
		HardwareCode:           deviceRequest.HardwareCode,
		IngCode:                deviceRequest.IngCode,
		LoggingCode:            deviceRequest.LoggingCode,
		SimSupported:           deviceRequest.SimSupported,
		SimType:                deviceRequest.SimType,
		Esim:                   strconv.FormatBool(deviceRequest.Esim),
		UserName:               userName,
	})

	return body, nil
}
