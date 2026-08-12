package job

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"backend/internal/biz"
	"backend/internal/conf"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-kratos/kratos/v2/log"
)

const depositOnlyBatchSize uint64 = 100

const buySomethingABI = `[
  {"inputs":[],"name":"getUserLength","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"internalType":"uint256","name":"startIndex","type":"uint256"},{"internalType":"uint256","name":"endIndex","type":"uint256"}],"name":"getUsersByIndex","outputs":[{"internalType":"address[]","name":"","type":"address[]"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"internalType":"uint256","name":"startIndex","type":"uint256"},{"internalType":"uint256","name":"endIndex","type":"uint256"}],"name":"getUsersAmountByIndex","outputs":[{"internalType":"uint256[]","name":"","type":"uint256[]"}],"stateMutability":"view","type":"function"}
]`

// DepositOnlyResult describes one contract-cursor synchronization.
type DepositOnlyResult struct {
	Contract   string `json:"contract"`
	Total      uint64 `json:"total"`
	CursorFrom uint64 `json:"cursor_from"`
	CursorTo   uint64 `json:"cursor_to"`
	Scanned    uint64 `json:"scanned"`
	Credited   uint64 `json:"credited"`
	Skipped    uint64 `json:"skipped"`
}

// ChainRechargeJob keeps the existing name for dependency compatibility. The
// recharge flow is intentionally request-driven through DepositOnly.
type ChainRechargeJob struct {
	walletRepo   biz.WalletRepo
	settingsRepo biz.SettingsRepo
	cfg          *conf.WalletConfig
	log          *log.Helper
	mu           sync.Mutex
}

func NewChainRechargeJob(
	walletRepo biz.WalletRepo,
	settingsRepo biz.SettingsRepo,
	cfg *conf.WalletConfig,
	logger log.Logger,
) *ChainRechargeJob {
	return &ChainRechargeJob{
		walletRepo: walletRepo, settingsRepo: settingsRepo,
		cfg: cfg, log: log.NewHelper(logger),
	}
}

// Start no longer starts an internal timer. An external scheduler calls the
// GET /api/admin_dhb/deposit_only endpoint once per minute.
func (j *ChainRechargeJob) Start() {
	j.log.Info("depositOnly recharge endpoint ready; internal chain scanner disabled")
}

func (j *ChainRechargeJob) Stop() {}

// DepositOnly mirrors new18new: use the BuySomething contract arrays as an
// append-only recharge ledger and only process entries after the saved cursor.
func (j *ChainRechargeJob) DepositOnly(ctx context.Context) (*DepositOnlyResult, error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.cfg == nil {
		return nil, fmt.Errorf("wallet config is missing")
	}
	rpcURL := strings.TrimSpace(j.cfg.GetRPCURL())
	contractRaw := strings.TrimSpace(j.cfg.GetDepositContract())
	if rpcURL == "" || !common.IsHexAddress(contractRaw) {
		return nil, fmt.Errorf("rpc or deposit contract is not configured")
	}

	contractAddress := common.HexToAddress(contractRaw)
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("connect RPC: %w", err)
	}
	defer client.Close()

	code, err := client.CodeAt(ctx, contractAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("read deposit contract: %w", err)
	}
	if len(code) == 0 {
		return nil, fmt.Errorf("deposit contract has no code")
	}

	parsedABI, err := abi.JSON(strings.NewReader(buySomethingABI))
	if err != nil {
		return nil, fmt.Errorf("parse deposit contract ABI: %w", err)
	}
	total, err := callUint256(ctx, client, parsedABI, contractAddress, "getUserLength")
	if err != nil {
		return nil, fmt.Errorf("getUserLength: %w", err)
	}
	if !total.IsUint64() {
		return nil, fmt.Errorf("deposit record count is too large")
	}

	totalCount := total.Uint64()
	cursorKey := depositOnlyCursorKey(contractAddress)
	cursor, err := j.loadCursor(ctx, cursorKey)
	if err != nil {
		return nil, err
	}
	if cursor > totalCount {
		return nil, fmt.Errorf("saved cursor %d exceeds contract length %d", cursor, totalCount)
	}

	result := &DepositOnlyResult{
		Contract: contractAddress.Hex(), Total: totalCount,
		CursorFrom: cursor, CursorTo: cursor,
	}
	for cursor < totalCount {
		end := cursor + depositOnlyBatchSize - 1
		if end >= totalCount {
			end = totalCount - 1
		}

		users, err := callAddressArray(ctx, client, parsedABI, contractAddress, "getUsersByIndex", cursor, end)
		if err != nil {
			return nil, fmt.Errorf("getUsersByIndex %d-%d: %w", cursor, end, err)
		}
		amounts, err := callUint256Array(ctx, client, parsedABI, contractAddress, "getUsersAmountByIndex", cursor, end)
		if err != nil {
			return nil, fmt.Errorf("getUsersAmountByIndex %d-%d: %w", cursor, end, err)
		}
		if len(users) != len(amounts) || len(users) != int(end-cursor+1) {
			return nil, fmt.Errorf("deposit contract returned mismatched arrays")
		}

		for i := range users {
			index := cursor + uint64(i)
			if amounts[i] == nil || amounts[i].Sign() <= 0 {
				result.Skipped++
				continue
			}
			credited, err := j.walletRepo.AutoCreditChainRecharge(
				ctx,
				depositOnlyRecordHash(contractAddress, index),
				users[i].Hex(),
				contractAddress.Hex(),
				amounts[i].String(),
				index,
			)
			if err != nil {
				return nil, fmt.Errorf("credit deposit index %d: %w", index, err)
			}
			if credited {
				result.Credited++
			} else {
				result.Skipped++
			}
		}

		cursor = end + 1
		if err := j.settingsRepo.Set(ctx, cursorKey, strconv.FormatUint(cursor, 10)); err != nil {
			return nil, fmt.Errorf("save deposit cursor: %w", err)
		}
		result.CursorTo = cursor
		result.Scanned = cursor - result.CursorFrom
	}

	j.log.Infof(
		"depositOnly synchronized: contract=%s total=%d from=%d to=%d credited=%d skipped=%d",
		result.Contract, result.Total, result.CursorFrom, result.CursorTo, result.Credited, result.Skipped,
	)
	return result, nil
}

func (j *ChainRechargeJob) loadCursor(ctx context.Context, key string) (uint64, error) {
	raw, err := j.settingsRepo.Get(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("read deposit cursor: %w", err)
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	cursor, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid deposit cursor %q", raw)
	}
	return cursor, nil
}

func callContract(ctx context.Context, client *ethclient.Client, contractABI abi.ABI, address common.Address, method string, args ...interface{}) ([]interface{}, error) {
	input, err := contractABI.Pack(method, args...)
	if err != nil {
		return nil, err
	}
	output, err := client.CallContract(ctx, ethereum.CallMsg{To: &address, Data: input}, nil)
	if err != nil {
		return nil, err
	}
	return contractABI.Unpack(method, output)
}

func callUint256(ctx context.Context, client *ethclient.Client, contractABI abi.ABI, address common.Address, method string) (*big.Int, error) {
	values, err := callContract(ctx, client, contractABI, address, method)
	if err != nil || len(values) != 1 {
		return nil, firstCallError(err, len(values))
	}
	value, ok := values[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("unexpected uint256 result")
	}
	return value, nil
}

func callAddressArray(ctx context.Context, client *ethclient.Client, contractABI abi.ABI, address common.Address, method string, start, end uint64) ([]common.Address, error) {
	values, err := callContract(ctx, client, contractABI, address, method, new(big.Int).SetUint64(start), new(big.Int).SetUint64(end))
	if err != nil || len(values) != 1 {
		return nil, firstCallError(err, len(values))
	}
	value, ok := values[0].([]common.Address)
	if !ok {
		return nil, fmt.Errorf("unexpected address[] result")
	}
	return value, nil
}

func callUint256Array(ctx context.Context, client *ethclient.Client, contractABI abi.ABI, address common.Address, method string, start, end uint64) ([]*big.Int, error) {
	values, err := callContract(ctx, client, contractABI, address, method, new(big.Int).SetUint64(start), new(big.Int).SetUint64(end))
	if err != nil || len(values) != 1 {
		return nil, firstCallError(err, len(values))
	}
	value, ok := values[0].([]*big.Int)
	if !ok {
		return nil, fmt.Errorf("unexpected uint256[] result")
	}
	return value, nil
}

func firstCallError(err error, values int) error {
	if err != nil {
		return err
	}
	return fmt.Errorf("unexpected result count %d", values)
}

func depositOnlyCursorKey(contract common.Address) string {
	return "deposit_only_cursor:" + strings.ToLower(contract.Hex())
}

func depositOnlyRecordHash(contract common.Address, index uint64) string {
	key := strings.ToLower(contract.Hex()) + ":" + strconv.FormatUint(index, 10)
	return crypto.Keccak256Hash([]byte(key)).Hex()
}
