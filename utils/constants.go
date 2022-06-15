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
var HomologationMustBeRegretion = "This homologation type is incorrect only can be a Regression."
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
var CREATE = "Create"
var PHASE = "Phase"
var EDIT = "Edit"
var CREATE_MAIN_MESSAGE = "The following Homologation process has been created on date"
var PLANNING_MAIN_MESSAGE = "The following Homologation process is planning to start on date"
var SAMPLE_START_MAIN_MESSAGE = "Samples has been received and are under revision on date"
var SAMPLE_END_MAIN_MESSAGE = "Samples are ready to start homologation process on date"
var TEST_MAIN_MESSAGE = "Testing process has been started on date"
var UNDER_MAIN_MESSAGE = "Testing process has been finished and is Under Decision Evaluation on date"
var FINISHED_MAIN_MESSAGE = "Homologation process has been finished and not apply carrier decision on date"
var APPROVED_MAIN_MESSAGE = "Homologation process has been APPROVED for the Carrier on date"
var REJECTED_MAIN_MESSAGE = "Homologation process has been REJECTED for the Carrier on date"
