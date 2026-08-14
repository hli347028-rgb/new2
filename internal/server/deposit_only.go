package server

import (
	"backend/internal/job"

	"github.com/go-kratos/kratos/v2/transport/http"
)

// RegisterDepositOnlyRoute exposes cron/HTTP triggers that accept a background
// poll cycle (default: 10 queries, 5s apart). Returns immediately so HTTP
// timeout (often 5s) does not cut the cycle short.
func RegisterDepositOnlyRoute(srv *http.Server, recharge *job.ChainRechargeJob, oracle *job.WinPriceOracleJob) {
	r := srv.Route("/")
	r.GET("/api/admin_dhb/deposit_only", func(ctx http.Context) error {
		return ctx.JSON(200, recharge.TriggerDepositOnlyCycle())
	})
	r.GET("/api/admin_dhb/deposit_only_win", func(ctx http.Context) error {
		return ctx.JSON(200, recharge.TriggerDepositOnlyWinCycle())
	})
	r.GET("/api/admin_dhb/win_price_oracle", func(ctx http.Context) error {
		return ctx.JSON(200, oracle.TriggerCycle())
	})
}
