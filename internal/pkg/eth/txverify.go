package eth

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
)

var erc20TransferEvent = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

// VerifyUSDTTransfer verifies an ERC20 USDT transfer in transaction logs.
// depositAddresses 可为多个平台收款地址，转账到其中任一地址即通过。
func VerifyUSDTTransfer(
	ctx context.Context,
	rpcURL, txHash, usdtContract string,
	depositAddresses []string,
	fromAddress string,
	expectedAmount decimal.Decimal,
	decimals int32,
) error {
	return VerifyERC20Transfer(ctx, rpcURL, txHash, usdtContract, depositAddresses, fromAddress, expectedAmount, decimals, "USDT")
}

// VerifyERC20Transfer 校验交易回执中是否存在指定 ERC20 合约的 Transfer 日志。
func VerifyERC20Transfer(
	ctx context.Context,
	rpcURL, txHash, tokenContract string,
	depositAddresses []string,
	fromAddress string,
	expectedAmount decimal.Decimal,
	decimals int32,
	tokenSymbol string,
) error {
	if rpcURL == "" {
		return nil
	}
	symbol := strings.TrimSpace(tokenSymbol)
	if symbol == "" {
		symbol = "token"
	}
	if tokenContract == "" {
		return fmt.Errorf("%s contract not configured", strings.ToLower(symbol))
	}
	allowed := make(map[string]struct{}, len(depositAddresses))
	for _, a := range depositAddresses {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		allowed[strings.ToLower(common.HexToAddress(a).Hex())] = struct{}{}
	}
	if len(allowed) == 0 {
		return fmt.Errorf("deposit address not configured")
	}

	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return fmt.Errorf("connect rpc failed: %w", err)
	}
	defer client.Close()

	hash := common.HexToHash(txHash)
	receipt, err := client.TransactionReceipt(ctx, hash)
	if err != nil {
		return fmt.Errorf("transaction receipt not found: %w", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("transaction failed")
	}

	expectedRaw, err := amountToRaw(expectedAmount, decimals)
	if err != nil {
		return err
	}

	contractAddr := common.HexToAddress(tokenContract)
	fromAddr := common.HexToAddress(fromAddress)

	for _, logItem := range receipt.Logs {
		if logItem == nil || len(logItem.Topics) < 3 {
			continue
		}
		if !strings.EqualFold(logItem.Address.Hex(), contractAddr.Hex()) {
			continue
		}
		if logItem.Topics[0] != erc20TransferEvent {
			continue
		}

		from := common.BytesToAddress(logItem.Topics[1].Bytes())
		to := common.BytesToAddress(logItem.Topics[2].Bytes())
		if !strings.EqualFold(from.Hex(), fromAddr.Hex()) {
			continue
		}
		if _, ok := allowed[strings.ToLower(to.Hex())]; !ok {
			continue
		}

		amount := new(big.Int).SetBytes(logItem.Data)
		if amount.Cmp(expectedRaw) < 0 {
			return fmt.Errorf("%s transfer amount insufficient", strings.ToLower(symbol))
		}
		return nil
	}
	return fmt.Errorf("%s transfer log not found", strings.ToLower(symbol))
}

func amountToRaw(amount decimal.Decimal, decimals int32) (*big.Int, error) {
	if decimals < 0 {
		return nil, fmt.Errorf("invalid token decimals")
	}
	multiplier := decimal.New(1, decimals)
	raw := amount.Mul(multiplier)
	if !raw.Equal(raw.Truncate(0)) {
		return nil, fmt.Errorf("amount precision exceeds token decimals")
	}
	return raw.BigInt(), nil
}
