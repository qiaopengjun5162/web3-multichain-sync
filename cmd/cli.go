package main

import (
	"context"
	"fmt"

	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/version"
	"github.com/urfave/cli/v2" // https://cli.urfave.org/v2/getting-started/

	multichain_sync_account "github.com/qiaopengjun5162/web3-multichain-sync"
	"github.com/qiaopengjun5162/web3-multichain-sync/common/cliapp"
	"github.com/qiaopengjun5162/web3-multichain-sync/common/opio"
	"github.com/qiaopengjun5162/web3-multichain-sync/config"
	"github.com/qiaopengjun5162/web3-multichain-sync/database"
	flags2 "github.com/qiaopengjun5162/web3-multichain-sync/flags"
	"github.com/qiaopengjun5162/web3-multichain-sync/notifier"
	"github.com/qiaopengjun5162/web3-multichain-sync/rpcclient"
	"github.com/qiaopengjun5162/web3-multichain-sync/rpcclient/chain-account/account"
	"github.com/qiaopengjun5162/web3-multichain-sync/services"
)

const (
	POLLING_INTERVAL     = 1 * time.Second
	MAX_RPC_MESSAGE_SIZE = 1024 * 1024 * 300
)

// Semantic holds the textual version string for major.minor.patch.
var Semantic = fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch)

// WithMeta holds the textual version string including the metadata.
var WithMeta = func() string {
	v := Semantic
	if version.Meta != "" {
		v += "-" + version.Meta
	}
	return v
}()

func withCommit(gitCommit, gitDate string) string {
	vsn := WithMeta
	if len(gitCommit) >= 8 {
		vsn += "-" + gitCommit[:8]
	}
	if (version.Meta != "stable") && (gitDate != "") {
		vsn += "-" + gitDate
	}
	return vsn
}

func runRpc(ctx *cli.Context, shutdown context.CancelCauseFunc) (cliapp.Lifecycle, error) {
	fmt.Println("running grpc server...")
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		log.Error("failed to load config", "err", err)
		return nil, err
	}
	grpcServerCfg := &services.BusinessMiddleConfig{
		GrpcHostname: cfg.RpcServer.Host,
		GrpcPort:     cfg.RpcServer.Port,
	}
	db, err := database.NewDB(ctx.Context, cfg.MasterDB)
	if err != nil {
		log.Error("failed to connect to database", "err", err)
		return nil, err
	}

	log.Info("Chain account rpc", "rpc uri", cfg.ChainAccountRpc)
	conn, err := grpc.NewClient(cfg.ChainAccountRpc, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Connect to da retriever fail", "err", err)
		return nil, err
	}
	client := account.NewWalletAccountServiceClient(conn)
	accountClient, err := rpcclient.NewWalletChainAccountClient(context.Background(), client, "Ethereum")
	if err != nil {
		log.Error("new wallet account client fail", "err", err)
		return nil, err
	}
	return services.NewBusinessMiddleWireServices(db, grpcServerCfg, accountClient)
}

func runMigrations(ctx *cli.Context) error {
	ctx.Context = opio.CancelOnInterrupt(ctx.Context)
	log.Info("running migrations...")
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		log.Error("failed to load config", "err", err)
		return err
	}
	db, err := database.NewDB(ctx.Context, cfg.MasterDB)
	if err != nil {
		log.Error("failed to connect to database", "err", err)
		return err
	}
	defer func(db *database.DB) {
		err := db.Close()
		if err != nil {
			log.Error("fail to close database", "err", err)
		}
	}(db)
	return db.ExecuteSQLMigration(cfg.Migrations)
}

func runMultichainSync(ctx *cli.Context, shutdown context.CancelCauseFunc) (cliapp.Lifecycle, error) {
	log.Info("exec wallet sync")
	cfg, err := config.LoadConfig(ctx)
	fmt.Println()
	if err != nil {
		log.Error("failed to load config", "err", err)
		return nil, err
	}
	return multichain_sync_account.NewMultiChainSync(ctx.Context, &cfg, shutdown)
}

func runNotify(ctx *cli.Context, shutdown context.CancelCauseFunc) (cliapp.Lifecycle, error) {
	fmt.Println("running notify task...")
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		log.Error("failed to load config", "err", err)
		return nil, err
	}
	db, err := database.NewDB(ctx.Context, cfg.MasterDB)
	if err != nil {
		log.Error("failed to connect to database", "err", err)
		return nil, err
	}
	return notifier.NewNotifier(db, shutdown)
}

func NewCli(GitCommit string, gitDate string) *cli.App {
	flags := flags2.Flags
	return &cli.App{
		Name:                 "Web3 multichain sync account",
		Usage:                "An exchange wallet scanner services with rpc and rest api services",
		Description:          "An exchange wallet scanner services with rpc and rest api services",
		Version:              withCommit(GitCommit, gitDate),
		EnableBashCompletion: true, // Boolean to enable bash completion commands
		Commands: []*cli.Command{
			{
				Name:        "rpc",
				Flags:       flags,
				Description: "Run rpc services",
				Action:      cliapp.LifecycleCmd(runRpc),
			},
			{
				Name:        "sync",
				Flags:       flags,
				Description: "Run rpc scanner wallet chain node",
				Action:      cliapp.LifecycleCmd(runMultichainSync),
			},
			{
				Name:        "notify",
				Flags:       flags,
				Description: "Run rpc scanner wallet chain node",
				Action:      cliapp.LifecycleCmd(runNotify),
			},
			{
				Name:        "migrate",
				Flags:       flags,
				Description: "Run database migrations",
				Action:      runMigrations,
			},
			{
				Name:        "version",
				Usage:       "Show project version",
				Description: "Show project version",
				Action: func(ctx *cli.Context) error {
					cli.ShowVersion(ctx)
					return nil
				},
			},
		},
	}
}
