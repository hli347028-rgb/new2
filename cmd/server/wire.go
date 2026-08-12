//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"backend/internal/biz"
	"backend/internal/conf"
	"backend/internal/data"
	"backend/internal/job"
	"backend/internal/server"
	"backend/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.DatabaseConfig, *conf.AuthConfig, *conf.WalletConfig, log.Logger) (*kratos.App, *job.SettlementJob, *job.ChainRechargeJob, *job.WinPriceOracleJob, *biz.AdminUsecase, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, job.ProviderSet, newApp))
}
