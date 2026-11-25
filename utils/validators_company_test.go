package utils_test

import (
	"reflect"
	"testing"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func TestNormalizeChileanRUT(t *testing.T) {
	valid := map[string]string{
		"12.345.678-5":  "12.345.678-5",
		"12345678-5":    "12.345.678-5",
		" 11.111.111-1": "11.111.111-1",
		"1.234.567-4":   "1.234.567-4",
		"1234567-4":     "1.234.567-4",
		"10.000.013-K":  "10.000.013-K",
		"10000013-k":    "10.000.013-K", // lowercase k
		"10000004-0":    "10.000.004-0",
	}
	for in, want := range valid {
		got, ok := utils.NormalizeChileanRUT(in)
		if !ok || got != want {
			t.Errorf("NormalizeChileanRUT(%q) = %q,%v want %q", in, got, ok, want)
		}
	}
	invalid := []string{
		"12.345.678-9",  // wrong check digit
		"10000013-0",    // K expected
		"10000004-K",    // 0 expected
		"123456785",     // no dash
		"12.345.678-",   // no check digit
		"12,345,678-5",  // commas
		"1.2345.678-5",  // bad grouping
		"AB.345.678-5",  // letters
		"123456789-5",   // too long
		"012345678-5",   // leading zero
		"12.345.678-55", // two check digits
		"",
	}
	for _, in := range invalid {
		if got, ok := utils.NormalizeChileanRUT(in); ok {
			t.Errorf("NormalizeChileanRUT(%q) accepted as %q", in, got)
		}
	}
}

func validCompany() *request.Company {
	return &request.Company{
		Name: "Xiaomi Multibanda", ClientID: " 110 ", RazonSocial: "Xiaomi Chile SpA",
		Address: "Av. Apoquindo 1234", Rut: "12345678-5",
	}
}

func TestValidateCompanyRequest(t *testing.T) {
	req := validCompany()
	if err := utils.ValidateCompanyRequest(req); err != nil {
		t.Fatal(err)
	}
	if req.Rut != "12.345.678-5" {
		t.Errorf("rut not normalized: %q", req.Rut)
	}

	cases := map[string]func(*request.Company){
		"Name is required":         func(r *request.Company) { r.Name = " " },
		"Client ID is required":    func(r *request.Company) { r.ClientID = "  " },
		"Razón social is required": func(r *request.Company) { r.RazonSocial = "" },
		"Address is required":      func(r *request.Company) { r.Address = "" },
		"RUT is required":          func(r *request.Company) { r.Rut = "" },
		"RUT is not valid: use the format 12.345.678-9 with a correct check digit": func(r *request.Company) { r.Rut = "12.345.678-9" },
	}
	for want, mutate := range cases {
		req := validCompany()
		mutate(req)
		err := utils.ValidateCompanyRequest(req)
		if !utils.IsValidationError(err) || err.Error() != want {
			t.Errorf("want %q, got %v", want, err)
		}
	}
}

func TestCompanyOabiMissing(t *testing.T) {
	if got := utils.CompanyOabiMissing("110", "12.345.678-5", "X SpA", "Calle 1"); got == nil || len(got) != 0 {
		t.Errorf("complete company: %v (must be an empty, non-nil list)", got)
	}
	got := utils.CompanyOabiMissing("", "12.345.678-5", " ", "")
	if !reflect.DeepEqual(got, []string{"client_id", "razon_social", "address"}) {
		t.Errorf("missing = %v", got)
	}
}
