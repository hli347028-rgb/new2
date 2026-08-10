package main

import (
	"flag"
	"os"
	"context"

	"backend/internal/conf"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z -X main.Name=backend"
var (
	// Name is the name of the compiled software.
	Name = "backend"
	// Version is the version of the compiled software.
	Version = "v1.0.0"
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	// 不依赖 Bootstrap protobuf 整包 Scan（server 段在部分环境下会扫成 nil）
	// 与 auth/wallet/database 一样按路径取值
	httpAddr, _ := c.Value("server.http.addr").String()
	grpcAddr, _ := c.Value("server.grpc.addr").String()
	if httpAddr == "" {
		httpAddr = "0.0.0.0:9000"
	}
	if grpcAddr == "" {
		grpcAddr = "0.0.0.0:9100"
	}
	serverCfg := &conf.Server{
		Http: &conf.Server_HTTP{Addr: httpAddr},
		Grpc: &conf.Server_GRPC{Addr: grpcAddr},
	}

	var authCfg conf.AuthConfig
	if err := c.Value("auth").Scan(&authCfg); err != nil {
		panic(err)
	}

	var walletCfg conf.WalletConfig
	if err := c.Value("wallet").Scan(&walletCfg); err != nil {
		panic(err)
	}

	var dbCfg conf.DatabaseConfig
	if err := c.Value("data.database").Scan(&dbCfg); err != nil {
		panic(err)
	}

	app, settlementJob, adminUsecase, cleanup, err := wireApp(serverCfg, &dbCfg, &authCfg, &walletCfg, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	_ = adminUsecase.LoadPersistedConfig(context.Background())

	settlementJob.Start()
	defer settlementJob.Stop()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
