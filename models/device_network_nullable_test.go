package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

// Network flags are nullable: a document without the field must decode as nil
// (not declared), while stored true/false values are kept as-is.
func TestDeviceNetworkFlagsDecodeMissingAsNil(t *testing.T) {
	raw, err := bson.Marshal(bson.M{
		"commercial_model": "X",
		"network_lte":      false,
		"network_gsm":      true,
		"band_5g_nsa":      []string{"n78"},
		"band_5g_sa":       []string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	var d Device
	if err := bson.Unmarshal(raw, &d); err != nil {
		t.Fatal(err)
	}
	if d.NetworkWcdma != nil || d.NetworkCarrierAgg != nil || d.Network5gSa != nil ||
		d.CommunicationRadio != nil {
		t.Fatalf("missing fields should decode as nil, got wcdma=%v ca=%v sa=%v radio=%v",
			d.NetworkWcdma, d.NetworkCarrierAgg, d.Network5gSa, d.CommunicationRadio)
	}
	if d.NetworkLte == nil || *d.NetworkLte {
		t.Fatalf("network_lte: want false, got %v", d.NetworkLte)
	}
	if d.NetworkGsm == nil || !*d.NetworkGsm {
		t.Fatalf("network_gsm: want true, got %v", d.NetworkGsm)
	}
	if len(d.Band5gNsa) != 1 || d.Band5gNsa[0] != "n78" {
		t.Fatalf("band_5g_nsa: got %v", d.Band5gNsa)
	}
	if d.Band5gSa == nil || len(d.Band5gSa) != 0 {
		t.Fatalf("band_5g_sa: want empty slice, got %v", d.Band5gSa)
	}
}
