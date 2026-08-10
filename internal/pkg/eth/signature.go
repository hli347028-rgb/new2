package eth

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// NormalizeAddress lowercases and ensures 0x prefix.
func NormalizeAddress(address string) (string, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return "", fmt.Errorf("address is empty")
	}
	if !common.IsHexAddress(address) {
		return "", fmt.Errorf("invalid hex address")
	}
	return strings.ToLower(common.HexToAddress(address).Hex()), nil
}

// VerifyPersonalSign verifies an Ethereum personal_sign signature.
func VerifyPersonalSign(message, signatureHex, expectedAddress string) error {
	expected, err := NormalizeAddress(expectedAddress)
	if err != nil {
		return err
	}

	signatureHex = strings.TrimPrefix(strings.TrimSpace(signatureHex), "0x")
	sig, err := hex.DecodeString(signatureHex)
	if err != nil {
		return fmt.Errorf("invalid signature hex: %w", err)
	}
	if len(sig) != 65 {
		return fmt.Errorf("invalid signature length")
	}
	if sig[64] >= 27 {
		sig[64] -= 27
	}

	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256Hash([]byte(prefix))

	pub, err := crypto.SigToPub(hash.Bytes(), sig)
	if err != nil {
		return fmt.Errorf("recover public key failed: %w", err)
	}
	recovered := strings.ToLower(crypto.PubkeyToAddress(*pub).Hex())
	if recovered != expected {
		return fmt.Errorf("signature address mismatch")
	}
	return nil
}
