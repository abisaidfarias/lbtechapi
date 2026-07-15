package services

import (
	"bytes"
	"fmt"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"github.com/xuri/excelize/v2"
)

func (s *multibandaService) ExportMultibanda(userID string) (bytes.Buffer, error) {
	items, err := s.Get(userID)
	var b bytes.Buffer
	if err != nil {
		return b, err
	}
	return exportMultibandaFile(items)
}

func exportMultibandaFile(items []*responses.MultibandaExpanded) (bytes.Buffer, error) {
	file := excelize.NewFile()
	for cell, header := range enums.ExcelMultibandaHeaders {
		file.SetCellValue(utils.PAGE, cell, header)
	}

	for index, m := range items {
		row := index + 2
		typeLabel := m.Type
		if label, ok := enums.MultibandaTypeLabels[m.Type]; ok {
			typeLabel = label
		}
		phaseLabel := fmt.Sprintf("%d", m.CurrentPhase)
		if label, ok := enums.MultibandaPhaseLabels[m.CurrentPhase]; ok {
			phaseLabel = label
		}
		statusLabel := fmt.Sprintf("%d", m.Status)
		if label, ok := enums.HomologationStatus_type[m.Status]; ok {
			statusLabel = label
		}
		osVersion := m.OsVersionView
		if osVersion == "" {
			osVersion = m.OsVersion
		}

		setCell := func(col int, value interface{}) {
			cell, _ := excelize.CoordinatesToCellName(col, row)
			file.SetCellValue(utils.PAGE, cell, value)
		}

		setCell(1, m.Company.Name)
		setCell(2, m.Brand.Name)
		setCell(3, m.Device.CommercialModel)
		setCell(4, m.Device.TechnicalModel)
		setCell(5, m.SoftwareVersion)
		setCell(6, m.HardwareVersion)
		setCell(7, osVersion)
		setCell(8, typeLabel)
		setCell(9, functions.MultibandaEvaluationTypeLabels(m.EvaluationType))
		setCell(10, phaseLabel)
		setCell(11, statusLabel)
		setCell(12, functions.ExcelDate(m.PlanningDate))
		setCell(13, functions.ExcelDate(m.SampleStartDate))
		setCell(14, functions.ExcelDate(m.SampleEndDate))
		setCell(15, functions.ExcelDate(m.TestStartDate))
		setCell(16, functions.ExcelDate(m.TestEndDate))
		setCell(17, functions.ExcelDate(m.UnderStartDate))
		setCell(18, functions.ExcelDate(m.UnderEndDate))
		setCell(19, functions.ExcelDate(m.CompletedDate))
		setCell(20, m.SubtelCertificateNumber)
		setCell(21, m.Comment)
		setCell(22, functions.ExcelYesNo(m.NeedReflash))
		setCell(23, m.CommentsReflash)
		setCell(24, functions.ExcelYesNo(m.IsInternalProject))
	}

	var b bytes.Buffer
	if err := file.Write(&b); err != nil {
		return b, err
	}
	return b, nil
}
