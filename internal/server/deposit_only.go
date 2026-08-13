package server

import (
	"backend/internal/job"

	"github.com/go-kratos/kratos/v2/transport/http"
)

// RegisterDepositOnlyRoute exposes idempotent contract-cursor recharge
// synchronization endpoints (also driven by the internal one-minute ticker).
func RegisterDepositOnlyRoute(srv *http.Server, recharge *job.ChainRechargeJob) {
	r := srv.Route("/")
	r.GET("/api/admin_dhb/deposit_only", func(ctx http.Context) error {
		result, err := recharge.DepositOnly(ctx)
		if err != nil {
			return err
		}
		return ctx.JSON(200, result)
	})
	r.GET("/api/admin_dhb/deposit_only_win", func(ctx http.Context) error {
		result, err := recharge.DepositOnlyWin(ctx)
		if err != nil {
			return err
		}
		return ctx.JSON(200, result)
	})
}
