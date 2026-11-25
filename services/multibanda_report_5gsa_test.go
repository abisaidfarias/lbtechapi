package services

import (
	"testing"

	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

func TestBuildReportPrefilledSupports5GSAOnlyWhenDeclaredTrue(t *testing.T) {
	yes, no := true, false
	cases := []struct {
		name string
		nsa  *bool
		sa   *bool
		want bool
	}{
		{"undeclared", nil, nil, false},
		{"sa false", nil, &no, false},
		{"nsa only", &yes, &no, false},
		{"nsa yes sa undeclared", &yes, nil, false},
		{"sa true", nil, &yes, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mb := &responses.MultibandaExpanded{}
			mb.Device.Network5gNsa = tc.nsa
			mb.Device.Network5gSa = tc.sa
			if got := buildReportPrefilled(mb).Supports5GSA; got != tc.want {
				t.Fatalf("Supports5GSA = %v, want %v", got, tc.want)
			}
		})
	}
}
