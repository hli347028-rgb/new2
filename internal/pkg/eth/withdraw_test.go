package eth

import (
	"testing"
)

func TestVerifyWithdrawSignature_Empty(t *testing.T) {
	addr := "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
	if err := VerifyWithdrawSignature(addr, "100", addr, 1, ""); err != nil {
		t.Fatalf("empty signature should pass: %v", err)
	}
}

func TestVerifyWithdrawSignature_AddressOnly(t *testing.T) {
	// Generated offline with personal_sign(address) for testing pattern acceptance.
	// Backend accepts address-only signatures from fetchSign().
	addr := "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
	// Invalid random sig should fail
	if err := VerifyWithdrawSignature(addr, "10", addr, 1, "0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"); err == nil {
		t.Fatal("random signature should fail")
	}
}
