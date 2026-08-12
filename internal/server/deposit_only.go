package server

import (
	"backend/internal/job"

	"github.com/go-kratos/kratos/v2/transport/http"
)

// RegisterDepositOnlyRoute exposes the idempotent contract-cursor recharge
// synchronization endpoint for an external one-minute scheduler.
func RegisterDepositOnlyRoute(srv *http.Server, recharge *job.ChainRechargeJob) {
	r := srv.Route("/")
	r.GET("/api/admin_dhb/deposit_only", func(ctx http.Context) error {
		result, err := recharge.DepositOnly(ctx)
		if err != nil {
			return err
		}
		return ctx.JSON(200, result)
	})
}
