package store

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
	"github.com/knadh/goyesql/v2"
)

type (
	PgOpts struct {
		Logg                 *slog.Logger
		DSN                  string
		MigrationsFolderPath string
		QueriesFolderPath    string
	}

	Pg struct {
		logg    *slog.Logger
		db      *pgxpool.Pool
		queries *queries
	}

	queries struct {
		RegisterName  string `query:"register-name"`
		UpdateName    string `query:"update-name"`
		UpsertName    string `query:"upsert-name"`
		LookupName    string `query:"lookup-name"`
		ReverseLookup string `query:"reverse-lookup"`
	}
)

func NewPgStore(o PgOpts) (Store, error) {
	parsedConfig, err := pgxpool.ParseConfig(o.DSN)
	if err != nil {
		return nil, err
	}

	dbPool, err := pgxpool.NewWithConfig(context.Background(), parsedConfig)
	if err != nil {
		return nil, err
	}

	queries, err := loadQueries(o.QueriesFolderPath)
	if err != nil {
		return nil, err
	}

	if err := runMigrations(context.Background(), dbPool, o.MigrationsFolderPath); err != nil {
		return nil, err
	}
	o.Logg.Info("migrations ran successfully")

	return &Pg{
		logg:    o.Logg,
		db:      dbPool,
		queries: queries,
	}, nil
}

func (pg *Pg) Close() {
	pg.db.Close()
}

func (pg *Pg) RegisterName(ctx context.Context, primaryName string, blockchainAddress string) error {
	_, err := pg.db.Exec(
		ctx,
		pg.queries.RegisterName,
		primaryName,
		blockchainAddress,
	)
	if err != nil {
		return err
	}

	return nil
}

func (pg *Pg) UpdateName(ctx context.Context, primaryName string, blockchainAddress string) error {
	_, err := pg.db.Exec(
		ctx,
		pg.queries.UpdateName,
		primaryName,
		blockchainAddress,
	)
	if err != nil {
		return err
	}

	return nil
}

func (pg *Pg) UpsertName(ctx context.Context, primaryName string, blockchainAddress string) error {
	_, err := pg.db.Exec(
		ctx,
		pg.queries.UpsertName,
		primaryName,
		blockchainAddress,
	)
	if err != nil {
		return err
	}

	return nil
}

func (pg *Pg) LookupName(ctx context.Context, primaryName string) (string, error) {
	var blockchainAddress string
	err := pg.db.QueryRow(
		ctx,
		pg.queries.LookupName,
		primaryName,
	).Scan(&blockchainAddress)
	if err != nil {
		return "", err
	}

	return blockchainAddress, nil
}

// TODO: If an address has multiple registered names, should we return them all? Should we allow this feature in the first place?
func (pg *Pg) ReverseLookup(ctx context.Context, blockchainAddress string) (string, error) {
	var primaryName string
	err := pg.db.QueryRow(
		ctx,
		pg.queries.ReverseLookup,
		blockchainAddress,
	).Scan(&primaryName)
	if err != nil {
		return "", err
	}

	return primaryName, nil
}

func loadQueries(queriesPath string) (*queries, error) {
	parsedQueries, err := goyesql.ParseFile(queriesPath)
	if err != nil {
		return nil, err
	}

	loadedQueries := &queries{}

	if err := goyesql.ScanToStruct(loadedQueries, parsedQueries, nil); err != nil {
		return nil, fmt.Errorf("failed to scan queries %v", err)
	}

	return loadedQueries, nil
}

func runMigrations(ctx context.Context, dbPool *pgxpool.Pool, migrationsPath string) error {
	const migratorTimeout = 15 * time.Second

	ctx, cancel := context.WithTimeout(ctx, migratorTimeout)
	defer cancel()

	conn, err := dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	migrator, err := migrate.NewMigrator(ctx, conn.Conn(), "schema_version")
	if err != nil {
		return err
	}

	if err := migrator.LoadMigrations(os.DirFS(migrationsPath)); err != nil {
		return err
	}

	if err := migrator.Migrate(ctx); err != nil {
		return err
	}

	return nil
}
