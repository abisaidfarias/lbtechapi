package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	validator "github.com/go-playground/validator/v10"
)

// TODO add correct error name message into gin response

// ValidPassword validates good password complexity
var ValidPassword validator.Func = func(fl validator.FieldLevel) bool {

	if err := verifyPassword(fl.Field().String()); err != nil {
		return false
	}

	return true
}

func verifyPassword(password string) error {
	var uppercasePresent bool
	var lowercasePresent bool
	var numberPresent bool
	var specialCharPresent bool
	const minPassLength = 8
	const maxPassLength = 64
	var passLen int
	var errorString string

	for _, ch := range password {
		switch {
		case unicode.IsNumber(ch):
			numberPresent = true
			passLen++
		case unicode.IsUpper(ch):
			uppercasePresent = true
			passLen++
		case unicode.IsLower(ch):
			lowercasePresent = true
			passLen++
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			specialCharPresent = true
			passLen++
		case ch == ' ':
			passLen++
		}
	}
	appendError := func(err string) {
		if len(strings.TrimSpace(errorString)) != 0 {
			errorString += ", " + err
		} else {
			errorString = err
		}
	}
	if !lowercasePresent {
		appendError("lowercase letter missing")
	}
	if !uppercasePresent {
		appendError("uppercase letter missing")
	}
	if !numberPresent {
		appendError("atleast one numeric character required")
	}
	if !specialCharPresent {
		appendError("special character missing")
	}
	if !(minPassLength <= passLen && passLen <= maxPassLength) {
		appendError(fmt.Sprintf("password length must be between %d to %d characters long", minPassLength, maxPassLength))
	}

	if len(errorString) != 0 {
		return fmt.Errorf(errorString)
	}
	return nil
}

// ValidTestCode validates correct code structure
var ValidTestCaseCode validator.Func = func(fl validator.FieldLevel) bool {
	codString := fl.Field().String()

	// TOOD remove fix optional check
	if len(codString) <= 0 {
		return true
	}

	incomingCode := fl.Field().String()

	var codeRegex = regexp.MustCompile(`^[A-Z]{8}\-[A-Z0-9]{2}[A-Z]$`)

	return codeRegex.MatchString(incomingCode)
}

func ValidateHomologationRequest(homologation *responses.Homologation,
	homologationRequest *request.Homologation) (bool, *models.CustomError) {
	if homologation == nil {
		if homologationRequest.Type != enums.HomologationType_value["INITIAL"] {
			var customeErr models.CustomError
			customeErr.Code = HomologationMustBeInitialCode
			customeErr.Err = HomologationMustBeInitial
			return false, &customeErr
		}
		return true, nil
	}
	if homologation.Status == enums.HomologationStatus_value["IN_PROGRESS"] {
		var customeErr models.CustomError
		customeErr.Code = HomologationExistCode
		customeErr.Err = HomologationExist
		return false, &customeErr
	}
	if homologation.Status == enums.HomologationStatus_value["APPROVED"] {
		if homologationRequest.Type != enums.HomologationType_value["MAINTENANCE"] {
			var customeErr models.CustomError
			customeErr.Code = HomologationMustBeMaintenanceCode
			customeErr.Err = HomologationMustBeMaintenance
			return false, &customeErr
		}
	}
	if homologation.Status == enums.HomologationStatus_value["REJECTED"] {
		if homologationRequest.Type != enums.HomologationType_value["REGRETION"] {
			var customeErr models.CustomError
			customeErr.Code = HomologationMustBeRegretionCode
			customeErr.Err = HomologationMustBeRegretion
			return false, &customeErr
		}
	}
	return true, nil

}
