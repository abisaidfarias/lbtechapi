package functions

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	utils "github.com/abisaidfarias/lbtechapi/utils"
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
func SendNotifications(to []string) {

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

	var body bytes.Buffer

	body.Write([]byte(fmt.Sprintf("Subject: This is a test subject \n%s\n\n", utils.MIME_HEADERS)))

	// Sending email.
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, body.Bytes())
	if err != nil {
		fmt.Println(err)
		return
	}

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
