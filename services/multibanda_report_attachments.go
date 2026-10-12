package services

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *multibandaReportService) UploadAttachment(
	multibandaID, userID, kind string,
	fileBytes []byte,
	originalFileName string,
) (*responses.MultibandaReportForm, error) {
	ctx, err := s.loadContext(multibandaID, userID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsInternal {
		return nil, fmt.Errorf("%w", utils.ErrorForbidden)
	}

	kind = strings.TrimSpace(strings.ToLower(kind))
	if err := utils.ValidateMultibandaReportAttachmentMeta(kind, originalFileName, int64(len(fileBytes))); err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(originalFileName)))
	suffix, downloadStem := attachmentSuffixAndStem(kind)
	if suffix == "" {
		return nil, utils.NewValidationError("invalid attachment kind")
	}

	technicalModel := strings.TrimSpace(ctx.prefilled.TechnicalModel)
	if technicalModel == "" {
		technicalModel = "model"
	}
	objectKey := multibandaReportAttachmentObjectKey(
		ctx.multibandaOID.Hex(), technicalModel, suffix, ext)
	downloadName := fmt.Sprintf("%s-%s-%s%s",
		ctx.multibandaOID.Hex(), technicalModel, downloadStem, ext)

	if s.storageService == nil {
		return nil, fmt.Errorf("storage not configured")
	}
	newURL, err := s.storageService.UploadFileWithKeyAndName(fileBytes, objectKey, downloadName)
	if err != nil {
		return nil, fmt.Errorf("upload attachment: %w", err)
	}

	oldURL := attachmentURL(ctx.report, kind)
	displayName := strings.TrimSpace(originalFileName)

	status := enums.MultibandaReportStatusDraft
	if ctx.report != nil && ctx.report.Status == enums.MultibandaReportStatusGenerated {
		status = enums.MultibandaReportStatusGenerated
	}

	req := mapping.MultibandaReportModelToSaveRequest(ctx.report)
	if req == nil {
		req = &request.MultibandaReportSave{}
	}
	applyAttachmentToSaveRequest(req, kind, newURL, displayName)
	report := mapping.MultibandaReportRequestToModel(
		req, ctx.multibandaOID, ctx.multibanda.Type, status)

	if err := s.reportRepository.Upsert(report); err != nil {
		return nil, err
	}
	if err := s.syncReportAttachmentsToProcess(ctx.multibandaOID, report); err != nil {
		return nil, err
	}
	if oldURL != "" && oldURL != newURL {
		_ = s.storageService.DeletePublicFile(oldURL)
	}

	return s.GetForm(multibandaID, userID)
}

func (s *multibandaReportService) syncReportAttachmentsToProcess(
	multibandaID primitive.ObjectID,
	report *models.MultibandaReport,
) error {
	if report == nil {
		return s.multibandaRepository.SetReportAttachments(multibandaID, "", "", "", "")
	}
	return s.multibandaRepository.SetReportAttachments(
		multibandaID,
		report.AdbInfoURL,
		report.AdbInfoOriginalFileName,
		report.BandValidationURL,
		report.BandValidationOriginalFileName,
	)
}

func attachmentSuffixAndStem(kind string) (s3Suffix, downloadStem string) {
	switch kind {
	case utils.MultibandaReportAttachmentAdbInfo:
		return "ADB-INFO", "ADB-INFO"
	case utils.MultibandaReportAttachmentBandValidation:
		return "BAND-VALIDATION", "BAND-VALIDATION"
	default:
		return "", ""
	}
}

func attachmentURL(report *models.MultibandaReport, kind string) string {
	if report == nil {
		return ""
	}
	switch kind {
	case utils.MultibandaReportAttachmentAdbInfo:
		return report.AdbInfoURL
	case utils.MultibandaReportAttachmentBandValidation:
		return report.BandValidationURL
	default:
		return ""
	}
}

func applyAttachmentToSaveRequest(req *request.MultibandaReportSave, kind, url, originalName string) {
	switch kind {
	case utils.MultibandaReportAttachmentAdbInfo:
		req.AdbInfoURL = url
		req.AdbInfoOriginalFileName = originalName
	case utils.MultibandaReportAttachmentBandValidation:
		req.BandValidationURL = url
		req.BandValidationOriginalFileName = originalName
	}
}
