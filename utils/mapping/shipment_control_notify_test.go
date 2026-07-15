package mapping

import (
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestShipmentControlToNotifyUsesShipmentCertificateNumbers(t *testing.T) {
	shipment := &models.ShipmentControl{
		SubtelCertificateNumber: "SUB-123",
		OabiCertificateNumber:   "OABI-456",
		ReworkNumber:            "22113344",
	}
	multibanda := &responses.MultibandaExpanded{
		ID:                      primitive.NewObjectID(),
		SubtelCertificateNumber: "",
	}

	notify := ShipmentControlToNotify(shipment, multibanda, "Xiaomi Multibanda", "Chile")

	if notify.SubtelCertificateNumber != "SUB-123" {
		t.Fatalf("subtel: got %q", notify.SubtelCertificateNumber)
	}
	if notify.OabiCertificate != "OABI-456" {
		t.Fatalf("oabi: got %q", notify.OabiCertificate)
	}
}

func TestShipmentControlToNotifyFallsBackToMultibandaSubtel(t *testing.T) {
	shipment := &models.ShipmentControl{
		PlanningDate: time.Now(),
	}
	multibanda := &responses.MultibandaExpanded{
		ID:                      primitive.NewObjectID(),
		SubtelCertificateNumber: "MB-999",
	}

	notify := ShipmentControlToNotify(shipment, multibanda, "Co", "CL")
	if notify.SubtelCertificateNumber != "MB-999" {
		t.Fatalf("subtel: got %q", notify.SubtelCertificateNumber)
	}
}

func TestShipmentControlToNotifyFallsBackToOabiCertificateLegacyField(t *testing.T) {
	shipment := &models.ShipmentControl{
		OabiCertificate: "LEGACY-OABI",
	}
	notify := ShipmentControlToNotify(shipment, nil, "Co", "CL")
	if notify.OabiCertificate != "LEGACY-OABI" {
		t.Fatalf("oabi: got %q", notify.OabiCertificate)
	}
}
