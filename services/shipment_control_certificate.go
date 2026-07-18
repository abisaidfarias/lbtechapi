package services

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *shipmentControlService) GenerateCertificate(
	id string,
	req *request.ShipmentControlCertificate,
	userID string,
) (*responses.ShipmentControlCertificate, error) {
	if err := s.requireProfileClaim(userID, enums.CanWriteShipmentControl); err != nil {
		return nil, err
	}

	shipmentControlID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, utils.NewValidationError("invalid shipment control id")
	}

	shipment, err := s.shipmentControlRepository.GetById(shipmentControlID)
	if err != nil {
		return nil, err
	}
	if shipment == nil {
		return nil, utils.NewValidationError("shipment control not found")
	}

	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if !functions.UserHasClientAccess(user, shipment.Company) {
		return nil, fmt.Errorf("%w", utils.ErrorForbidden)
	}

	if shipment.CurrentPhase != enums.ShipmentControlPhaseUnderRevision {
		return nil, utils.NewValidationError("certificate can only be generated during Under Revision phase")
	}

	registroOABI := strings.TrimSpace(req.RegistroOABI)
	alreadyGenerated := strings.TrimSpace(shipment.OabiCertificateUrl) != ""

	company, err := s.companyRepository.GetById(shipment.Company.Hex())
	if err != nil {
		return nil, err
	}
	if company == nil {
		return nil, utils.NewValidationError("company not found")
	}
	if err := validateCompanyCertificateFields(company); err != nil {
		return nil, err
	}

	multibanda, err := s.multibandaRepository.GetByIdExpanded(shipment.Multibanda)
	if err != nil {
		return nil, err
	}
	if multibanda == nil {
		return nil, utils.NewValidationError("multibanda not found")
	}

	controlNumber := functions.ControlNumberFromCertificateURL(shipment.OabiCertificateUrl)
	if controlNumber == "" {
		controlNumber, err = s.nextShipmentControlNumber(company.ClientID)
		if err != nil {
			return nil, err
		}
	}

	certData := functions.BuildShipmentControlCertificateData(
		company,
		multibanda,
		controlNumber,
		req.RegistroOABI,
		req.RegisteredCount(),
		shipment.ReworkNumber,
	)

	htmlBytes, err := functions.RenderShipmentControlCertificateHTML(
		certData,
		utils.TEMPLATE_SHIPMENT_CONTROL_CERTIFICATE_PATH,
	)
	if err != nil {
		return nil, fmt.Errorf("render certificate html: %w", err)
	}

	pdfBytes, err := shipmentControlCertificateHTMLToPDF(htmlBytes)
	if err != nil {
		return nil, fmt.Errorf("generate certificate pdf: %w", err)
	}

	objectKey := shipmentCertificateObjectKey(controlNumber)
	certificateURL, err := s.uploadShipmentCertificateToS3(objectKey, pdfBytes)
	if err != nil {
		return nil, err
	}

	if err := s.shipmentControlRepository.UpdateCertificate(shipmentControlID, certificateURL, registroOABI); err != nil {
		return nil, err
	}

	return &responses.ShipmentControlCertificate{
		URL:           functions.CacheBustCertificateURL(certificateURL),
		RegistroOABI:  registroOABI,
		ControlNumber: controlNumber,
		Regenerated:   alreadyGenerated,
	}, nil
}

func validateCompanyCertificateFields(company *responses.Company) error {
	if strings.TrimSpace(company.ClientID) == "" {
		return utils.NewValidationError("company client_id is required to generate certificate")
	}
	if strings.TrimSpace(company.Rut) == "" {
		return utils.NewValidationError("company rut is required to generate certificate")
	}
	if strings.TrimSpace(company.RazonSocial) == "" && strings.TrimSpace(company.Name) == "" {
		return utils.NewValidationError("company razon_social is required to generate certificate")
	}
	return nil
}

func (s *shipmentControlService) nextShipmentControlNumber(clientID string) (string, error) {
	now := time.Now()
	prefix := functions.ShipmentControlControlNumberPrefix(clientID, now)
	count, err := s.shipmentControlRepository.CountCertificatesByControlPrefix(prefix)
	if err != nil {
		return "", err
	}
	return functions.BuildShipmentControlControlNumber(clientID, int(count)+1, now), nil
}

func shipmentCertificateObjectKey(controlNumber string) string {
	prefix := strings.TrimSpace(os.Getenv("SHIPMENT_CERT_S3_PREFIX"))
	if prefix == "" {
		prefix = "shipment-control-certificates"
	}
	prefix = strings.Trim(prefix, "/")
	safeName := strings.NewReplacer("/", "-", "\\", "-", " ", "-").Replace(controlNumber)
	return fmt.Sprintf("%s/%s.pdf", prefix, safeName)
}

func (s *shipmentControlService) uploadShipmentCertificateToS3(objectKey string, pdfBytes []byte) (string, error) {
	if s.storageService == nil {
		return "", fmt.Errorf("storage not configured")
	}
	if len(pdfBytes) == 0 {
		return "", fmt.Errorf("empty pdf bytes")
	}

	var url string
	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		url, err = s.storageService.UploadFileWithKey(pdfBytes, objectKey)
		if err == nil {
			return url, nil
		}
		log.Printf("shipment certificate s3 upload attempt %d/3 (key=%s): %v", attempt, objectKey, err)
		if attempt < 3 {
			time.Sleep(time.Duration(attempt) * 400 * time.Millisecond)
		}
	}
	return "", fmt.Errorf("upload certificate to s3: %w", err)
}
