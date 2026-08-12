package job

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestDepositOnlyCursorKeyIsContractSpecific(t *testing.T) {
	a := common.HexToAddress("0x1111111111111111111111111111111111111111")
	b := common.HexToAddress("0x2222222222222222222222222222222222222222")
	if depositOnlyCursorKey(a) == depositOnlyCursorKey(b) {
		t.Fatal("cursor key must be contract-specific")
	}
}

func TestDepositOnlyRecordHashIsStableAndUnique(t *testing.T) {
	contract := common.HexToAddress("0xe11c2F7902CB03cAA38F80B27DC20702af14D5c7")
	one := depositOnlyRecordHash(contract, 1)
	oneAgain := depositOnlyRecordHash(contract, 1)
	two := depositOnlyRecordHash(contract, 2)
	if one != oneAgain {
		t.Fatal("same contract index must produce the same record hash")
	}
	if one == two {
		t.Fatal("different indexes must produce unique record hashes")
	}
	if len(one) != 66 {
		t.Fatalf("record hash must fit tx_hash, got length %d", len(one))
	}
}
