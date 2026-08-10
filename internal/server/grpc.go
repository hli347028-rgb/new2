package server

import (
	authv1 "backend/api/auth/v1"
	walletv1 "backend/api/wallet/v1"
	"backend/internal/conf"
	authmw "backend/internal/middleware"
	"backend/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, auth *service.AuthService, wallet *service.WalletService, _ log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			authmw.BearerToken(),
		),
	}
	if c == nil {
		c = &conf.Server{}
	}
	if c.Grpc == nil {
		c.Grpc = &conf.Server_GRPC{Addr: "0.0.0.0:9100"}
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	authv1.RegisterAuthServer(srv, auth)
	walletv1.RegisterWalletServer(srv, wallet)
	return srv
}
