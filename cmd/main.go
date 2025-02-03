package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/grassrootseconomics/ens-offchain-resolver/internal/api"
	"github.com/grassrootseconomics/ens-offchain-resolver/internal/store"
	"github.com/grassrootseconomics/ens-offchain-resolver/internal/util"
	"github.com/grassrootseconomics/ens-offchain-resolver/pkg/ens"
	"github.com/knadh/koanf/v2"
)

const defaultGracefulShutdownPeriod = time.Second * 20

var (
	build = "dev"

	confFlag             string
	migrationsFolderFlag string
	queriesFlag          string

	lo *slog.Logger
	ko *koanf.Koanf
)

func init() {
	flag.StringVar(&confFlag, "config", "config.toml", "Config file location")
	flag.StringVar(&migrationsFolderFlag, "migrations", "migrations/", "Migrations folder location")
	flag.StringVar(&queriesFlag, "queries", "queries.sql", "Queries file location")
	flag.Parse()

	lo = util.InitLogger()
	ko = util.InitConfig(lo, confFlag)

	lo.Info("starting resolver service", "build", build)
}

func main() {
	var wg sync.WaitGroup
	ctx, stop := notifyShutdown()

	publicKey, err := util.LoadSigningKey(ko.MustString("api.public_key"))
	if err != nil {
		lo.Error("could not load JWT verifying key", "error", err)
		os.Exit(1)
	}

	chainSigner, err := crypto.HexToECDSA(ko.MustString("chain.signer_private_key"))
	if err != nil {
		lo.Error("could not load chain signer private key", "error", err)
		os.Exit(1)
	}

	store, err := store.NewPgStore(store.PgOpts{
		Logg:                 lo,
		DSN:                  ko.MustString("postgres.dsn"),
		MigrationsFolderPath: migrationsFolderFlag,
		QueriesFolderPath:    queriesFlag,
	})
	if err != nil {
		lo.Error("could not initialize postgres store", "error", err)
		os.Exit(1)
	}

	apiServer := api.New(api.APIOpts{
		VerifyingKey:  publicKey,
		EnableMetrics: ko.Bool("metrics.enable"),
		ListenAddress: ko.MustString("api.address"),
		Store:         store,
		Logg:          lo,
		ENSProvider:   ens.NewProvider(chainSigner),
	})

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := apiServer.Start(); err != nil {
			lo.Error("failed to start HTTP server", "err", fmt.Sprintf("%T", err))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	lo.Info("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultGracefulShutdownPeriod)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := apiServer.Stop(shutdownCtx); err != nil {
			lo.Error("failed to stop HTTP server", "err", fmt.Sprintf("%T", err))
		}
	}()

	go func() {
		wg.Wait()
		stop()
		cancel()
		os.Exit(0)
	}()

	<-shutdownCtx.Done()
	if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
		stop()
		cancel()
		lo.Error("graceful shutdown period exceeded, forcefully shutting down")
	}
	os.Exit(1)
}

func notifyShutdown() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
}

// func loadQueries(queriesPath string) (*data.PgQueries, error) {
// 	// parsedQueries, err := goyesql.ParseFile(queriesPath)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	// // loadedQueries := &data.PgQueries{}

// 	// if err := goyesql.ScanToStruct(loadedQueries, parsedQueries, nil); err != nil {
// 	// 	return nil, fmt.Errorf("failed to scan queries %v", err)
// 	// }

// 	// return loadedQueries, nil
// }
