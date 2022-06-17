package functions

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"time"

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

	var JWTKey = []byte(os.Getenv("SECRET_KEY"))
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

	// var toList []string
	// if !isInternal {
	// 	notificationCollection := database.GetInstance().Collection("notifications")

	// 	var notification *responses.Notification
	// 	err := notificationCollection.FindOne(context.TODO(),
	// 		queries.GetNotifictionByCompany(companyId)).Decode(&notification)
	// 	if err != nil {
	// 		return
	// 	}
	// 	for _, user := range notification.NotificationEmails {
	// 		toList = append(toList, user.Email)
	// 	}
	// } else {
	// 	userCollection := database.GetInstance().Collection("users")

	// 	var user *responses.UserExpanded

	// 	cursor, err := userCollection.Aggregate(context.TODO(), queries.GetUserExpandedInternal())
	// 	if err != nil {
	// 		return
	// 	}
	// 	for cursor.Next(context.TODO()) {
	// 		cursor.Decode(&user)
	// 		toList = append(toList, user.Email)
	// 	}
	// }
	// if len(toList) == 0 {
	// 	return
	// }
	from := os.Getenv("EMAIL_FROM")
	password := os.Getenv("EMAIL_PASSWORD")

	smtpHost := os.Getenv("SMTP_CLIENTE")
	smtpPort := os.Getenv("EMAIL_PORT")

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
	approvalType int, carrier string, testingType string, planningDate string,
	sampleStartDate string, sampleEndDate string, testStartDate string,
	testEndDate string, underStartDate string, underEndDate string,
	resultDate string, templatePath string) (bytes.Buffer, error) {

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
		ApprovalType:    enums.HomologationType_key[approvalType],
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
	})

	return body, nil
}
func GetTrackingBodyMessage(subject string, mainMessge string, client string,
	brand string, technicalModel string, commercialModel string,
	responsible string, externalResponsible string, country string,
	location string, imeis string, comment string, templatePath string) (bytes.Buffer, error) {

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
		MainMessage         string
		Date                string
		Client              string
		Brand               string
		TechnicalModel      string
		CommercialModel     string
		Responsible         string
		ExternalResponsible string
		Country             string
		Location            string
		IMEI                string
		Comments            string
	}{
		MainMessage:         mainMessge,
		Date:                fmt.Sprintf("%02d/%02d/%d", time.Now().Day(), time.Now().Month(), time.Now().Year()),
		Client:              client,
		Brand:               brand,
		TechnicalModel:      technicalModel,
		CommercialModel:     commercialModel,
		Responsible:         responsible,
		ExternalResponsible: externalResponsible,
		Country:             country,
		Location:            location,
		IMEI:                imeis,
		Comments:            comment,
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
			return nil, errors.New("Unknown from server")
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
			subject = fmt.Sprintf("Subject: Homologation Process Finishes %s %s %s",
				country, brand, commercialModel)
			return utils.FINISHED_MAIN_MESSAGE, subject
		}
	}
	return mainMessage, subject
}
