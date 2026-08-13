package job

import (
	"context"
	"testing"
	"time"

	"backend/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
)

func TestFetchWinPriceLive(t *testing.T) {
	j := NewWinPriceOracleJob(nil, &conf.WalletConfig{
		RPCURL:                       "https://rpc1.eoeo.info",
		UsdtContract:                 "0x926632975149221891f1b9B56Efd125Dfe90ba2f",
		WinPair:                      "0x15ad085fc866370b59936575565434b14d22281d",
		WinPriceOracleEnabled:        true,
		WinPricePollSeconds:          60,
		WinPriceQueriesPerCycle:      10,
		WinPriceQueryIntervalSeconds: 5,
	}, log.DefaultLogger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	price, err := j.fetchWinPriceUSDT(ctx)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if price <= 0 {
		t.Fatalf("unexpected price %v", price)
	}
	t.Logf("1 WIN = %v USDT", price)
}
