package mapping

import (
	"testing"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestEnrichShipmentControlSubtelCertificateUsesMultibandaFallback(t *testing.T) {
	item := &responses.ShipmentControlExpanded{}
	multibanda := &responses.MultibandaExpanded{
		ID:                      primitive.NewObjectID(),
		SubtelCertificateNumber: "1950980",
	}

	EnrichShipmentControlSubtelCertificate(item, multibanda)

	if item.SubtelCertificateNumber != "1950980" {
		t.Fatalf("root subtel: got %q", item.SubtelCertificateNumber)
	}
	if item.Multibanda.SubtelCertificateNumber != "1950980" {
		t.Fatalf("multibanda subtel: got %q", item.Multibanda.SubtelCertificateNumber)
	}
}

func TestPreserveShipmentControlSubtelCertificateKeepsExisting(t *testing.T) {
	shipment := &models.ShipmentControl{}
	existing := &models.ShipmentControl{SubtelCertificateNumber: "SC-1"}
	multibanda := &responses.MultibandaExpanded{SubtelCertificateNumber: "MB-1"}

	PreserveShipmentControlSubtelCertificate(shipment, existing, multibanda)

	if shipment.SubtelCertificateNumber != "SC-1" {
		t.Fatalf("subtel: got %q", shipment.SubtelCertificateNumber)
	}
}

func TestPreserveShipmentControlSubtelCertificateUsesMultibandaWhenMissing(t *testing.T) {
	shipment := &models.ShipmentControl{}
	multibanda := &responses.MultibandaExpanded{SubtelCertificateNumber: "MB-1"}

	PreserveShipmentControlSubtelCertificate(shipment, nil, multibanda)

	if shipment.SubtelCertificateNumber != "MB-1" {
		t.Fatalf("subtel: got %q", shipment.SubtelCertificateNumber)
	}
}
