package mapping

import (
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func PrinterRequestToPrinter(printer *request.Printer, percentageTonner string) *models.Printer {

	// isInternal, _ := strconv.ParseBool()
	var detail models.Detail
	detail.Pages = printer.Pages
	detail.Location = printer.Location
	detail.MaxTonner = printer.MaxToner
	detail.RemToner = printer.RemToner
	detail.SNconsumible = printer.SNconsumible
	detail.PNconsumible = printer.PNconsumible
	detail.CreatedDate = time.Now()
	detail.PercentageTonner = percentageTonner

	var details []models.Detail
	details = append(details, detail)

	return &models.Printer{
		Modelo:           printer.Modelo,
		Serial:           printer.Serial,
		Pages:            printer.Pages,
		Location:         printer.Location,
		MaxTonner:        printer.MaxToner,
		RemToner:         printer.RemToner,
		SNconsumible:     printer.SNconsumible,
		PNconsumible:     printer.PNconsumible,
		CreatedDate:      time.Now(),
		PercentageTonner: percentageTonner,
		Details:          details,
	}
}
func PrinterToDetail(printer *models.Printer) models.Detail {

	return models.Detail{
		Pages:            printer.Pages,
		Location:         printer.Location,
		MaxTonner:        printer.MaxTonner,
		RemToner:         printer.RemToner,
		SNconsumible:     printer.SNconsumible,
		PNconsumible:     printer.PNconsumible,
		PercentageTonner: printer.PercentageTonner,
		CreatedDate:      printer.CreatedDate,
	}
}
