package functions

import (
	"fmt"
	"time"

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
