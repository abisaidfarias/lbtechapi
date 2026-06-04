package functions

import (
	"sort"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ApplyShipmentControlPhaseDateRules(
	shipmentControl *models.ShipmentControl,
	existing *models.ShipmentControl,
) {
	if shipmentControl == nil {
		return
	}

	switch shipmentControl.CurrentPhase {
	case enums.ShipmentControlPhaseValidation:
		if shipmentControl.ValidationStartDate.IsZero() {
			if existing != nil && !existing.ValidationStartDate.IsZero() {
				shipmentControl.ValidationStartDate = existing.ValidationStartDate
			} else if existing != nil && !existing.PlanningDate.IsZero() {
				shipmentControl.ValidationStartDate = existing.PlanningDate
			}
		}
	case enums.ShipmentControlPhaseUnderRevision:
		validationEnd := shipmentControl.ValidationEndDate
		if validationEnd.IsZero() && existing != nil {
			validationEnd = existing.ValidationEndDate
		}
		if !validationEnd.IsZero() && shipmentControl.UnderRevisionStartDate.IsZero() {
			shipmentControl.UnderRevisionStartDate = validationEnd
		}
	case enums.ShipmentControlPhaseCompleted:
		if shipmentControl.CompletedDate.IsZero() {
			if existing != nil && !existing.CompletedDate.IsZero() {
				shipmentControl.CompletedDate = existing.CompletedDate
			} else {
				underRevisionEnd := shipmentControl.UnderRevisionEndDate
				if underRevisionEnd.IsZero() && existing != nil {
					underRevisionEnd = existing.UnderRevisionEndDate
				}
				if !underRevisionEnd.IsZero() {
					shipmentControl.CompletedDate = underRevisionEnd
				}
			}
		}
	}
}

func ApplyShipmentControlStatusRules(
	shipmentControl *models.ShipmentControl,
	existing *models.ShipmentControl,
) {
	if shipmentControl == nil {
		return
	}

	underRevisionEnd := shipmentControl.UnderRevisionEndDate
	if underRevisionEnd.IsZero() && existing != nil {
		underRevisionEnd = existing.UnderRevisionEndDate
	}

	if !underRevisionEnd.IsZero() {
		shipmentControl.CurrentPhase = enums.ShipmentControlPhaseCompleted
		shipmentControl.Status = enums.ShipmentControlStatusCompleted
		return
	}

	shipmentControl.Status = enums.ShipmentControlStatusInProgress
}

func GroupAvailableMultibandas(
	company responses.Company,
	multibandas []*responses.MultibandaExpanded,
) *responses.ShipmentControlAvailableResponse {
	grouped := make(map[primitive.ObjectID]*responses.ShipmentControlAvailableDevice)
	order := make([]primitive.ObjectID, 0, len(multibandas))

	for _, multibanda := range multibandas {
		if multibanda == nil {
			continue
		}
		deviceID := multibanda.Device.ID
		entry, ok := grouped[deviceID]
		if !ok {
			entry = &responses.ShipmentControlAvailableDevice{
				Device:  multibanda.Device,
				Brand:   multibanda.Brand,
				Options: []responses.ShipmentControlAvailableOption{},
			}
			grouped[deviceID] = entry
			order = append(order, deviceID)
		}

		entry.Options = append(entry.Options, responses.ShipmentControlAvailableOption{
			MultibandaID:      multibanda.ID.Hex(),
			SubtelCertificateNumber: multibanda.SubtelCertificateNumber,
			SoftwareVersion:   multibanda.SoftwareVersion,
			HardwareVersion:   multibanda.HardwareVersion,
			OsVersion:         multibanda.OsVersion,
			OsVersionView:     multibanda.OsVersionView,
		})
	}

	sort.Slice(order, func(i, j int) bool {
		left := grouped[order[i]].Device.CommercialModel
		right := grouped[order[j]].Device.CommercialModel
		return left < right
	})

	devices := make([]responses.ShipmentControlAvailableDevice, 0, len(order))
	for _, deviceID := range order {
		devices = append(devices, *grouped[deviceID])
	}

	return &responses.ShipmentControlAvailableResponse{
		Company: company,
		Devices: devices,
	}
}

func UserHasClientAccess(user *responses.User, companyID primitive.ObjectID) bool {
	if user == nil {
		return false
	}
	if !user.IsInternal {
		return user.Company == companyID
	}
	if len(user.Clients) == 0 {
		return true
	}
	for _, clientID := range user.Clients {
		if clientID == companyID {
			return true
		}
	}
	return false
}
