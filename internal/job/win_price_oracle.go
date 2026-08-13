package job

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"backend/internal/biz"
	"backend/internal/conf"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

// Uniswap V2 Pair: getReserves / token0 / token1
const uniswapV2PairABI = `[
  {"constant":true,"inputs":[],"name":"getReserves","outputs":[
    {"internalType":"uint112","name":"_reserve0","type":"uint112"},
    {"internalType":"uint112","name":"_reserve1","type":"uint112"},
    {"internalType":"uint32","name":"_blockTimestampLast","type":"uint32"}
  ],"payable":false,"stateMutability":"view","type":"function"},
  {"constant":true,"inputs":[],"name":"token0","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},
  {"constant":true,"inputs":[],"name":"token1","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"}
]`

// WinPriceOracleJob 定时从 WinSwap V2 Pair 读取 WWIN/USDT 储备并写入 WIN 价格。
type WinPriceOracleJob struct {
	admin *biz.AdminUsecase
	cfg   *conf.WalletConfig
	log   *log.Helper
	stop  chan struct{}
	mu    sync.Mutex
}

func NewWinPriceOracleJob(admin *biz.AdminUsecase, cfg *conf.WalletConfig, logger log.Logger) *WinPriceOracleJob {
	return &WinPriceOracleJob{
		admin: admin,
		cfg:   cfg,
		log:   log.NewHelper(logger),
		stop:  make(chan struct{}),
	}
}

func (j *WinPriceOracleJob) Start() {
	if j.cfg == nil || !j.cfg.IsWinPriceOracleEnabled() {
		j.log.Info("win price oracle disabled")
		return
	}
	go j.run()
	j.log.Infof("win price oracle started: pair=%s cycle=%ds queries=%d interval=%ds rpc=%s",
		j.cfg.GetWinPair(),
		j.cfg.GetWinPricePollSeconds(),
		j.cfg.GetWinPriceQueriesPerCycle(),
		j.cfg.GetWinPriceQueryIntervalSeconds(),
		j.cfg.GetRPCURL())
}

func (j *WinPriceOracleJob) Stop() {
	select {
	case <-j.stop:
	default:
		close(j.stop)
	}
}

// run：每分钟一轮，每轮连续查询 10 次，相邻间隔 5 秒（可用配置覆盖）。
func (j *WinPriceOracleJob) run() {
	for {
		select {
		case <-j.stop:
			return
		default:
		}

		cycleStart := time.Now()
		queries := int(j.cfg.GetWinPriceQueriesPerCycle())
		if queries <= 0 {
			queries = 10
		}
		gap := time.Duration(j.cfg.GetWinPriceQueryIntervalSeconds()) * time.Second
		if gap <= 0 {
			gap = 5 * time.Second
		}

		for i := 0; i < queries; i++ {
			select {
			case <-j.stop:
				return
			default:
			}
			j.pollOnce()
			if i >= queries-1 {
				break
			}
			select {
			case <-time.After(gap):
			case <-j.stop:
				return
			}
		}

		cycle := time.Duration(j.cfg.GetWinPricePollSeconds()) * time.Second
		if cycle <= 0 {
			cycle = time.Minute
		}
		if remain := cycle - time.Since(cycleStart); remain > 0 {
			select {
			case <-time.After(remain):
			case <-j.stop:
				return
			}
		}
	}
}

func (j *WinPriceOracleJob) pollOnce() {
	j.mu.Lock()
	defer j.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	price, err := j.fetchWinPriceUSDT(ctx)
	if err != nil {
		j.log.Errorf("win price oracle fetch failed: %v", err)
		return
	}
	if err := j.admin.SetWinPriceFromOracle(ctx, price); err != nil {
		j.log.Errorf("win price oracle persist failed: %v", err)
		return
	}
	j.log.Infof("win price oracle updated: 1 WIN = %s USDT", decimal.NewFromFloat(price).StringFixed(8))
}

func (j *WinPriceOracleJob) fetchWinPriceUSDT(ctx context.Context) (float64, error) {
	rpcURL := j.cfg.GetRPCURL()
	pairRaw := j.cfg.GetWinPair()
	usdtRaw := strings.TrimSpace(j.cfg.GetUsdtContract())
	if rpcURL == "" || !common.IsHexAddress(pairRaw) {
		return 0, fmt.Errorf("rpc or win_pair not configured")
	}
	if usdtRaw == "" || !common.IsHexAddress(usdtRaw) {
		return 0, fmt.Errorf("usdt_contract not configured")
	}

	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return 0, fmt.Errorf("dial rpc: %w", err)
	}
	defer client.Close()

	pair := common.HexToAddress(pairRaw)
	usdt := common.HexToAddress(usdtRaw)
	parsedABI, err := abi.JSON(strings.NewReader(uniswapV2PairABI))
	if err != nil {
		return 0, fmt.Errorf("parse pair abi: %w", err)
	}

	token0, err := callAddress(ctx, client, parsedABI, pair, "token0")
	if err != nil {
		return 0, fmt.Errorf("token0: %w", err)
	}
	token1, err := callAddress(ctx, client, parsedABI, pair, "token1")
	if err != nil {
		return 0, fmt.Errorf("token1: %w", err)
	}
	reserve0, reserve1, err := callGetReserves(ctx, client, parsedABI, pair)
	if err != nil {
		return 0, fmt.Errorf("getReserves: %w", err)
	}
	if reserve0.Sign() <= 0 || reserve1.Sign() <= 0 {
		return 0, fmt.Errorf("empty reserves: r0=%s r1=%s", reserve0.String(), reserve1.String())
	}

	var winReserve, usdtReserve *big.Int
	switch {
	case strings.EqualFold(token0.Hex(), usdt.Hex()):
		usdtReserve, winReserve = reserve0, reserve1
	case strings.EqualFold(token1.Hex(), usdt.Hex()):
		winReserve, usdtReserve = reserve0, reserve1
	default:
		return 0, fmt.Errorf("pair tokens do not include usdt_contract: token0=%s token1=%s usdt=%s",
			token0.Hex(), token1.Hex(), usdt.Hex())
	}

	// 两边同为 18 decimals 时，原始储备比值即为 USDT/WIN
	priceDec := decimal.NewFromBigInt(usdtReserve, 0).Div(decimal.NewFromBigInt(winReserve, 0))
	price, _ := priceDec.Float64() // exact=false 仅表示浮点精度损耗，属正常
	if price <= 0 {
		return 0, fmt.Errorf("invalid computed price: %s", priceDec.String())
	}
	return price, nil
}

func callAddress(ctx context.Context, client *ethclient.Client, contractABI abi.ABI, address common.Address, method string) (common.Address, error) {
	values, err := callContract(ctx, client, contractABI, address, method)
	if err != nil || len(values) != 1 {
		return common.Address{}, firstCallError(err, len(values))
	}
	value, ok := values[0].(common.Address)
	if !ok {
		return common.Address{}, fmt.Errorf("unexpected address result")
	}
	return value, nil
}

func callGetReserves(ctx context.Context, client *ethclient.Client, contractABI abi.ABI, address common.Address) (*big.Int, *big.Int, error) {
	input, err := contractABI.Pack("getReserves")
	if err != nil {
		return nil, nil, err
	}
	output, err := client.CallContract(ctx, ethereum.CallMsg{To: &address, Data: input}, nil)
	if err != nil {
		return nil, nil, err
	}
	values, err := contractABI.Unpack("getReserves", output)
	if err != nil || len(values) < 2 {
		return nil, nil, firstCallError(err, len(values))
	}
	r0, ok0 := values[0].(*big.Int)
	r1, ok1 := values[1].(*big.Int)
	if !ok0 || !ok1 {
		return nil, nil, fmt.Errorf("unexpected reserves type")
	}
	return r0, r1, nil
}
