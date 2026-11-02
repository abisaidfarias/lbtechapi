// Diagnostic: go run ./tools/test_imei_info from repo root (loads .env).
package main

import (
	"context"
	"fmt"

	_ "github.com/abisaidfarias/lbtechapi/config"
	"github.com/abisaidfarias/lbtechapi/services"
	imeiutil "github.com/abisaidfarias/lbtechapi/utils/imei"
)

func main() {
	client := services.NewImeiInfoClient()
	cases := []struct {
		label string
		imei  string
	}{
		{"valid_pdf_example", "350105094917765"},
		{"invalid_luhn", "350105094917764"},
	}

	for _, c := range cases {
		fmt.Printf("\n=== %s ===\n", c.label)
		n := imeiutil.Normalize(c.imei)
		fmt.Printf("imei=%s luhn=%v tac=%s\n", n, imeiutil.IsValidLuhn(n), imeiutil.TAC(n))
		if !imeiutil.IsValidLuhn(n) {
			fmt.Println("OK: would reject before imei.info")
			continue
		}
		res, err := client.Lookup(context.Background(), n)
		if err != nil {
			fmt.Printf("FAIL: %v\n", err)
			continue
		}
		fmt.Printf("OK brand=%q technical=%q commercial=%q\n",
			res.Brand, res.TechnicalModel, res.CommercialName)
	}
}
