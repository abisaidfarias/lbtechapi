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

func (s *shipmentControlService) ExportShipmentControl(userID string) (bytes.Buffer, error) {
	items, err := s.Get(userID)
	var b bytes.Buffer
	if err != nil {
		return b, err
	}
	return exportShipmentControlFile(items)
}

func exportShipmentControlFile(items []*responses.ShipmentControlExpanded) (bytes.Buffer, error) {
	file := excelize.NewFile()
	for cell, header := range enums.ExcelShipmentControlHeaders {
		file.SetCellValue(utils.PAGE, cell, header)
	}

	for index, sc := range items {
		row := index + 2
		phaseLabel := fmt.Sprintf("%d", sc.CurrentPhase)
		if label, ok := enums.ShipmentControlPhaseLabels[sc.CurrentPhase]; ok {
			phaseLabel = label
		}
		statusLabel := fmt.Sprintf("%d", sc.Status)
		if label, ok := enums.ShipmentControlStatusLabels[sc.Status]; ok {
			statusLabel = label
		}
		osVersion := sc.Multibanda.OsVersionView
		if osVersion == "" {
			osVersion = sc.Multibanda.OsVersion
		}
		subtelNumber := sc.SubtelCertificateNumber
		if subtelNumber == "" {
			subtelNumber = sc.Multibanda.SubtelCertificateNumber
		}

		setCell := func(col int, value interface{}) {
			cell, _ := excelize.CoordinatesToCellName(col, row)
			file.SetCellValue(utils.PAGE, cell, value)
		}

		setCell(1, sc.Client)
		setCell(2, sc.Company.Name)
		setCell(3, sc.Country.Name)
		setCell(4, sc.Device.Brand)
		setCell(5, sc.Device.CommercialModel)
		setCell(6, sc.Device.TechnicalModel)
		setCell(7, sc.Device.PlatformOs)
		setCell(8, sc.Multibanda.SoftwareVersion)
		setCell(9, sc.Multibanda.HardwareVersion)
		setCell(10, osVersion)
		setCell(11, subtelNumber)
		setCell(12, phaseLabel)
		setCell(13, statusLabel)
		setCell(14, functions.ExcelDate(sc.PlanningDate))
		setCell(15, functions.ExcelDate(sc.ValidationStartDate))
		setCell(16, functions.ExcelDate(sc.ValidationEndDate))
		setCell(17, functions.ExcelDate(sc.UnderRevisionStartDate))
		setCell(18, functions.ExcelDate(sc.UnderRevisionEndDate))
		setCell(19, functions.ExcelDate(sc.CompletedDate))
		setCell(20, sc.ImeiQuantity)
		setCell(21, sc.RegisteredImeiCount)
		setCell(22, sc.ReworkNumber)
		setCell(23, sc.OabiCertificateNumber)
		setCell(24, sc.Comment)
	}

	var b bytes.Buffer
	if err := file.Write(&b); err != nil {
		return b, err
	}
	return b, nil
}
