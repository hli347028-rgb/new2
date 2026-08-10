package eth

import (
	"fmt"
	"strings"
)

// VerifyWithdrawSignature validates withdraw signatures from taurus frontend.
// Accepted patterns:
//  1. empty signature — rely on Bearer JWT (withdrawal page sends empty person.sign)
//  2. personal_sign(wallet address) — fetchSign() default
//  3. personal_sign(full withdraw message) — legacy structured message
func VerifyWithdrawSignature(userAddress, amount, toAddress string, withdrawAt int64, signature string) error {
	if strings.TrimSpace(signature) == "" {
		return nil
	}

	normalized, err := NormalizeAddress(userAddress)
	if err != nil {
		return err
	}

	if err := VerifyPersonalSign(normalized, signature, normalized); err == nil {
		return nil
	}

	message := fmt.Sprintf(
		"Withdraw USDT from backend account\nAddress: %s\nAmount: %s USDT\nTo: %s\nWithdrawAt: %d",
		normalized, amount, toAddress, withdrawAt,
	)
	return VerifyPersonalSign(message, signature, normalized)
}
