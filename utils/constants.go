package utils

import "errors"

// ErrorInvalidCredentials Invalid credentials"
var BaseUrlAzureBlob = "https://%s.blob.core.windows.net/%s"

// ErrorInvalidURLParams invalid url params
var ImageContainer = "images"

//ERROR CONSTANTS
var HomologationExist = "This homologation already exist please verify"
var HomologationExistCode = 100
var HomologationMustBeInitial = "This homologation type is incorrect only can be a Initial."
var HomologationMustBeInitialCode = 101
var HomologationMustBeMaintenance = "This homologation type is incorrect only can be a Maintenance."
var HomologationMustBeMaintenanceCode = 102
var HomologationMustBeRegretion = "This homologation type is incorrect only can be a Regretion."
var HomologationMustBeRegretionCode = 103
var OtherCategory = "Others"
var PAGE = "Sheet1"
var COMPANY = "Company"
var SOFTWARE_VERSION = "Sw Version"
var TECHNICAL_MODEL = "Technical Model"
var BRAND = "Brand"
var COUNTRY = "Country"
var MIME_HEADERS = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
var ErrorInvalidCredentials = errors.New("invalid credentials")
