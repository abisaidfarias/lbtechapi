package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/services/pdfengine"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	certificateWorkerAttempts  = 2
	certificateWorkerRetryWait = 5 * time.Second
	certificateWorkerSlack     = 30 * time.Second
	certificateFailedUserError = "certificate generation failed, please retry"
	certificateErrorMaxLen     = 300
)

// certificateWorkerTimeout is the budget for the whole job. It is derived from
// the render budget so the two cannot drift apart: a worker deadline shorter
// than attempts x render kills the job mid-render and always reports failure.
func certificateWorkerTimeout() time.Duration {
	render := pdfengine.RenderTimeout()
	return time.Duration(certificateWorkerAttempts)*render +
		time.Duration(certificateWorkerAttempts-1)*certificateWorkerRetryWait +
		certificateWorkerSlack
}

// certificateStaleJobAfter must outlive the worker, otherwise a user retrying
// from the modal claims a job that is still rendering and the two pile up
// behind the render semaphore.
func certificateStaleJobAfter() time.Duration {
	return certificateWorkerTimeout() + time.Minute
}

func (s *shipmentControlService) GenerateCertificate(
	id string,
	req *request.ShipmentControlCertificate,
	userID string,
) (*responses.ShipmentControlCertificateAccepted, error) {
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
		return nil, utils.NewValidationError("certificate only available in Under Revision phase")
	}

	registroOABI := strings.TrimSpace(req.RegistroOABI)
	if registroOABI == "" {
		return nil, utils.NewValidationError("missing field: registro_oabi")
	}
	registeredCount := req.RegisteredCount()
	if registeredCount <= 0 {
		return nil, utils.NewValidationError("missing field: registered_imei_count")
	}

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

	controlNumber := functions.ControlNumberFromCertificateURL(shipment.OabiCertificateUrl)
	if controlNumber == "" && shipment.OabiCertificateState != nil {
		controlNumber = strings.TrimSpace(shipment.OabiCertificateState.ControlNumber)
	}
	if controlNumber == "" {
		controlNumber, err = s.nextShipmentControlNumber(company.ClientID)
		if err != nil {
			return nil, err
		}
	}

	now := time.Now()
	claimState := models.OabiCertificateState{
		Status:        models.OabiCertificateStatusGenerating,
		ControlNumber: controlNumber,
		StartedAt:     now,
		InputSnapshot: models.OabiCertificateInputSnapshot{
			RegistroOABI:        registroOABI,
			RegisteredImeiCount: registeredCount,
		},
	}

	staleBefore := now.Add(-certificateStaleJobAfter())
	claimed, err := s.shipmentControlRepository.ClaimCertificateGeneration(shipmentControlID, claimState, staleBefore)
	if err != nil {
		return nil, err
	}
	if !claimed {
		return nil, utils.ErrCertificateGenerating
	}

	s.generateCertificateAsync(shipmentControlID.Hex())

	return &responses.ShipmentControlCertificateAccepted{
		Status:     models.OabiCertificateStatusGenerating,
		ShipmentID: shipmentControlID.Hex(),
	}, nil
}

func (s *shipmentControlService) GetCertificateStatus(
	id string,
	userID string,
) (*responses.ShipmentControlCertificateStatus, error) {
	if err := s.requireProfileClaim(userID, enums.CanReadShipmentControl); err != nil {
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

	return mapCertificateStatus(shipment.OabiCertificateState), nil
}

func mapCertificateStatus(state *models.OabiCertificateState) *responses.ShipmentControlCertificateStatus {
	if state == nil || strings.TrimSpace(state.Status) == "" {
		return &responses.ShipmentControlCertificateStatus{Status: "none"}
	}

	resp := &responses.ShipmentControlCertificateStatus{
		Status:        state.Status,
		URL:           state.URL,
		ControlNumber: state.ControlNumber,
	}
	if !state.GeneratedAt.IsZero() {
		resp.GeneratedAt = state.GeneratedAt.UTC().Format(time.RFC3339)
	}
	if state.Status == models.OabiCertificateStatusFailed {
		msg := strings.TrimSpace(state.Error)
		if msg == "" {
			msg = certificateFailedUserError
		}
		resp.Error = msg
	}
	return resp
}

func (s *shipmentControlService) generateCertificateAsync(shipmentControlID string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic generating certificate %s: %v", shipmentControlID, r)
				s.markCertificateFailed(shipmentControlID, certificateFailureMessage(fmt.Errorf("panic: %v", r)))
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), certificateWorkerTimeout())
		defer cancel()

		var lastErr error
		for attempt := 1; attempt <= certificateWorkerAttempts; attempt++ {
			if lastErr = s.generateCertificateOnce(ctx, shipmentControlID); lastErr == nil {
				return
			}
			log.Printf("shipment certificate %s attempt %d/%d failed: %v",
				shipmentControlID, attempt, certificateWorkerAttempts, lastErr)
			if attempt < certificateWorkerAttempts {
				time.Sleep(certificateWorkerRetryWait)
			}
		}
		log.Printf("shipment certificate %s abandoned after %d attempts: %v",
			shipmentControlID, certificateWorkerAttempts, lastErr)
		s.markCertificateFailed(shipmentControlID, certificateFailureMessage(lastErr))
	}()
}

func (s *shipmentControlService) generateCertificateOnce(ctx context.Context, shipmentControlID string) error {
	oid, err := primitive.ObjectIDFromHex(shipmentControlID)
	if err != nil {
		return err
	}

	shipment, err := s.shipmentControlRepository.GetById(oid)
	if err != nil {
		return err
	}
	if shipment == nil || shipment.OabiCertificateState == nil {
		return fmt.Errorf("shipment certificate state not found")
	}
	if shipment.OabiCertificateState.Status != models.OabiCertificateStatusGenerating {
		return nil
	}

	snapshot := shipment.OabiCertificateState.InputSnapshot
	controlNumber := strings.TrimSpace(shipment.OabiCertificateState.ControlNumber)
	if controlNumber == "" {
		controlNumber = functions.ControlNumberFromCertificateURL(shipment.OabiCertificateUrl)
	}
	if controlNumber == "" {
		return fmt.Errorf("missing control number for certificate generation")
	}

	company, err := s.companyRepository.GetById(shipment.Company.Hex())
	if err != nil || company == nil {
		return fmt.Errorf("company not found for certificate generation")
	}

	multibanda, err := s.multibandaRepository.GetByIdExpanded(shipment.Multibanda)
	if err != nil || multibanda == nil {
		return fmt.Errorf("multibanda not found for certificate generation")
	}

	certData := functions.BuildShipmentControlCertificateData(
		company,
		multibanda,
		controlNumber,
		snapshot.RegistroOABI,
		snapshot.RegisteredImeiCount,
		shipment.ReworkNumber,
	)

	htmlBytes, err := functions.RenderShipmentControlCertificateHTML(
		certData,
		utils.TEMPLATE_SHIPMENT_CONTROL_CERTIFICATE_PATH,
	)
	if err != nil {
		return fmt.Errorf("render certificate html: %w", err)
	}

	if s.pdfEngine == nil {
		return fmt.Errorf("pdf engine not configured")
	}
	pdfBytes, err := s.pdfEngine.RenderPDF(ctx, string(htmlBytes))
	if err != nil {
		return fmt.Errorf("generate certificate pdf: %w", err)
	}

	objectKey := shipmentCertificateObjectKey(controlNumber)
	baseURL, err := s.uploadShipmentCertificateToS3(objectKey, pdfBytes)
	if err != nil {
		return err
	}

	certificateURL := functions.CacheBustCertificateURL(baseURL)
	generatedAt := time.Now()
	return s.shipmentControlRepository.MarkCertificateReady(
		oid,
		certificateURL,
		strings.TrimSpace(snapshot.RegistroOABI),
		generatedAt,
	)
}

// certificateFailureMessage turns the internal error into something that
// explains the failure in the modal. Storing one generic string for every cause
// made timeouts, a missing Chrome and S3 errors indistinguishable without
// pulling EB logs.
func certificateFailureMessage(err error) string {
	if err == nil {
		return certificateFailedUserError
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return certificateFailedUserError
	}

	lower := strings.ToLower(msg)
	var hint string
	switch {
	case errors.Is(err, context.DeadlineExceeded), strings.Contains(lower, "deadline exceeded"):
		hint = "PDF render timed out (raise SHIPMENT_CERT_PDF_TIMEOUT_SEC or instance CPU)"
	case strings.Contains(lower, "chrome failed to start"),
		strings.Contains(lower, "executable file not found"):
		hint = "Chrome/Chromium unavailable (install it or set CHROME_PATH)"
	case strings.Contains(lower, "upload certificate to s3"):
		hint = "could not upload the certificate to storage"
	default:
		hint = msg
	}
	return truncateRunes(hint, certificateErrorMaxLen)
}

func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}

func (s *shipmentControlService) markCertificateFailed(shipmentControlID, message string) {
	oid, err := primitive.ObjectIDFromHex(shipmentControlID)
	if err != nil {
		log.Printf("certificate failed invalid id %s: %v", shipmentControlID, err)
		return
	}
	if err := s.shipmentControlRepository.MarkCertificateFailed(oid, message); err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			log.Printf("certificate mark failed %s: %v", shipmentControlID, err)
		}
	}
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
