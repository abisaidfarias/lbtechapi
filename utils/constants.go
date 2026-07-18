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
var ErrCertificateGenerating = errors.New("certificate generation in progress")
var CREATE = "Create"
var PHASE = "Phase"
var REQUEST_DELETE = "RequestDelete"
var DELETED = "Deleted"
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
var MULTIBANDA_CREATE_MAIN_MESSAGE = "The following Multibanda process has been created on date"
var MULTIBANDA_PLANNING_MAIN_MESSAGE = "The following Multibanda process is planning to start on date"
var MULTIBANDA_SAMPLE_START_MAIN_MESSAGE = "Samples have been received and are under revision on date"
var MULTIBANDA_SAMPLE_END_MAIN_MESSAGE = "Samples are ready to start the Multibanda process on date"
var MULTIBANDA_TEST_MAIN_MESSAGE = "Testing process has been started on date"
var MULTIBANDA_UNDER_MAIN_MESSAGE = "Testing process has been finished and is under decision evaluation on date"
var MULTIBANDA_FINISHED_MAIN_MESSAGE = "Multibanda process has been finished and does not apply Laboratory decision on date"
var MULTIBANDA_APPROVED_MAIN_MESSAGE = "Multibanda process has been APPROVED by the Laboratory on date"
var MULTIBANDA_REJECTED_MAIN_MESSAGE = "Multibanda process has been REJECTED by the Laboratory on date"
var TEMPLATE_TRACKING_MOVE_PATH = "utils/htmlMessageTemplate/device_tracking_move.html"
var TEMPLATE_MULTIBANDA_PHASE_PATH = "utils/htmlMessageTemplate/multibanda_phase.html"
var SHIPMENT_CONTROL_CREATE_MAIN_MESSAGE = "The following Shipment Control request has been created on date"
var SHIPMENT_CONTROL_VALIDATION_START_MAIN_MESSAGE = "The Shipment Control validation process has started on date"
var SHIPMENT_CONTROL_VALIDATION_END_MAIN_MESSAGE = "The Shipment Control validation process has ended on date"
var SHIPMENT_CONTROL_COMPLETE_MAIN_MESSAGE = "The Shipment Control process has been completed on date"
var SHIPMENT_CONTROL_REQUEST_DELETE_INTERNAL_MAIN_MESSAGE = "The client %s has requested the deletion of the following Shipment Control project on date"
var SHIPMENT_CONTROL_REQUEST_DELETE_CLIENT_MAIN_MESSAGE = "We have received your request to delete the following Shipment Control project on date. You will receive a notification once the project has been deleted"
var SHIPMENT_CONTROL_DELETED_MAIN_MESSAGE = "The following Shipment Control project has been deleted on date"
var TEMPLATE_SHIPMENT_CONTROL_PHASE_PATH = "utils/htmlMessageTemplate/shipment_control_phase.html"
var TEMPLATE_SHIPMENT_CONTROL_CERTIFICATE_PATH = "utils/htmlMessageTemplate/shipment_control_certificate.html"
var TEMPLATE_HOMOLOGATION_PATH = "utils/htmlMessageTemplate/createHomologation.html"
var TEMPLATE_FAIL_PATH = "utils/htmlMessageTemplate/createFail.html"
var CREATE_FAIL_MAIN_MESSAGE = "A new issue has been created on date"
var CREATE_DEVICE_MAIN_MESSAGE = "The following Device has been created on date"
var UPDATE_DEVICE_MAIN_MESSAGE = "The following Device has been updated on date"
var TEMPLATE_DEVICE_PATH = "utils/htmlMessageTemplate/device.html"

// MaxUploadFileSize is the maximum allowed multipart upload size (50 MiB).
const MaxUploadFileSize int64 = 50 << 20
