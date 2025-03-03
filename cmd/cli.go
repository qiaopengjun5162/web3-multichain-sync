package main

import (
	"fmt"
	"time"

	"github.com/qiaopengjun5162/web3-multichain-sync/database"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/version"
	"github.com/urfave/cli/v2" // https://cli.urfave.org/v2/getting-started/

	"github.com/qiaopengjun5162/web3-multichain-sync/common/opio"
	"github.com/qiaopengjun5162/web3-multichain-sync/config"
	flags2 "github.com/qiaopengjun5162/web3-multichain-sync/flags"
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
