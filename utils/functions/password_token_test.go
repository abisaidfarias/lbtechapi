package functions

import "testing"

func TestHashPasswordTokenDeterministic(t *testing.T) {
	raw := "sample-token-value"
	a := HashPasswordToken(raw)
	b := HashPasswordToken(raw)
	if a != b || a == "" {
		t.Fatalf("expected stable non-empty hash")
	}
}

func TestGeneratePasswordTokenUnique(t *testing.T) {
	raw1, hash1, err := GeneratePasswordToken()
	if err != nil {
		t.Fatal(err)
	}
	raw2, hash2, err := GeneratePasswordToken()
	if err != nil {
		t.Fatal(err)
	}
	if raw1 == raw2 || hash1 == hash2 {
		t.Fatal("expected unique tokens")
	}
}
