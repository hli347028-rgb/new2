package server

import (
	adminv1 "backend/api/admin/v1"
	authv1 "backend/api/auth/v1"
	walletv1 "backend/api/wallet/v1"
	"backend/internal/conf"
	"backend/internal/job"
	authmw "backend/internal/middleware"
	"backend/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, auth *service.AuthService, wallet *service.WalletService, admin *service.AdminService, legacy *service.AdminLegacyService, recharge *job.ChainRechargeJob, oracle *job.WinPriceOracleJob, _ log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Filter(authmw.CORS()),
		http.Middleware(
			recovery.Recovery(),
			authmw.BearerToken(),
		),
	}
	if c == nil {
		c = &conf.Server{}
	}
	if c.Http == nil {
		c.Http = &conf.Server_HTTP{Addr: "0.0.0.0:9000"}
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	authv1.RegisterAuthHTTPServer(srv, auth)
	walletv1.RegisterWalletHTTPServer(srv, wallet)
	adminv1.RegisterAdminHTTPServer(srv, admin)
	service.RegisterWalletExtraRoutes(srv, wallet)
	RegisterAdminLegacyRoutes(srv, legacy)
	RegisterDepositOnlyRoute(srv, recharge, oracle)
	return srv
}
