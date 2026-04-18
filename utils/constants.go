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
var TRACKING_MOVE = "MOVEMENT"
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
var CREATE_TRACKING_MAIN_MESSAGE = "New Sample(s) arrives to our Laboratory/Office"
var MOVE_TRACKING_MAIN_MESSAGE = "Following Sample(s) has been moved of location"
var TEMPLATE_TRACKING_PATH = "utils/htmlMessageTemplate/device_tracking.html"
var TEMPLATE_TRACKING_MOVE_PATH = "utils/htmlMessageTemplate/device_tracking_move.html"
var TEMPLATE_HOMOLOGATION_PATH = "utils/htmlMessageTemplate/createHomologation.html"
var TEMPLATE_FAIL_PATH = "utils/htmlMessageTemplate/createFail.html"
var CREATE_FAIL_MAIN_MESSAGE = "A new issue has been created on date"
var CREATE_DEVICE_MAIN_MESSAGE = "The following Device has been created on date"
var UPDATE_DEVICE_MAIN_MESSAGE = "The following Device has been updated on date"
var TEMPLATE_DEVICE_PATH = "utils/htmlMessageTemplate/device.html"
