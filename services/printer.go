package services

import (
	"fmt"
	"strconv"

	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// IPrinterService is the printer service
type IPrinterService interface {
	Create(*request.Printer) error
	Get() ([]*responses.Printer, error)
}

type printerService struct {
	printerRepository repositories.IPrinterRepository
}

// NewPrinterService is a constructor
func NewPrinterService(printerRepository repositories.IPrinterRepository) IPrinterService {
	return &printerService{
		printerRepository: printerRepository,
	}
}

// Create creates a new cateogry
func (s *printerService) Create(printerRequest *request.Printer) error {

	printer := mapping.PrinterRequestToPrinter(printerRequest)
	remTonner, _ := strconv.ParseFloat(printer.RemToner, 64)
	maxTonner, _ := strconv.ParseFloat(printer.MaxTonner, 64)
	totalDiv := (remTonner / maxTonner) * 100
	printer.PercentageTonner = fmt.Sprintf("%f", totalDiv) + "%"

	err := s.printerRepository.Create(printer)

	if err != nil {
		return err
	}

	return nil
}

// Get gets a list of all categories
func (s *printerService) Get() ([]*responses.Printer, error) {
	result, err := s.printerRepository.Get()

	if err != nil {
		return nil, err
	}

	return result, nil
}
