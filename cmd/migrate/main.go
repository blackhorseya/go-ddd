// Package main is the database migration entry point for the Identity bounded
// context. It runs in two modes from a single binary:
//
//   - As a CLI, with the subcommands below.
//   - As an AWS Lambda function, when AWS_LAMBDA_RUNTIME_API is present. Only
//     "up" is reachable that way; see handleMigrate.
//
// Usage:
//
//	migrate [-config path] [-yes] <command> [args]
//
// Commands:
//
//	up             apply every pending migration
//	down           roll back every applied migration (requires -yes)
//	steps <n>      move n migrations forward (n > 0) or backward (n < 0)
//	version        print the current schema version
//	force <v>      pin the version and clear the dirty flag, without running SQL
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/blackhorseya/go-ddd/internal/identity/infrastructure/persistence"
	"github.com/blackhorseya/go-ddd/internal/shared/infrastructure/config"
)

// lambdaRuntimeAPIEnv is set by AWS only inside the Lambda execution
// environment, which is how the binary knows which mode to start in.
const lambdaRuntimeAPIEnv = "AWS_LAMBDA_RUNTIME_API"

func main() {
	if os.Getenv(lambdaRuntimeAPIEnv) != "" {
		lambda.Start(handleMigrate)

		return
	}

	if err := runCLI(); err != nil {
		log.Fatalf("migrate: %v", err)
	}
}

// migrateResult reports where the schema stands once the function returns.
type migrateResult struct {
	Version uint   `json:"version"`
	Dirty   bool   `json:"dirty"`
	Message string `json:"message"`
}

// handleMigrate applies every pending migration.
//
// It deliberately accepts no event payload: only "up" is reachable from Lambda,
// because the -yes confirmation guarding the destructive commands has no
// equivalent in an invocation. Configuration comes from APP_* environment
// variables — there is no config file in the deployment package.
//
// The context is unused today because golang-migrate's API takes none; it stays
// in the signature so database work can be scoped to the invocation later.
func handleMigrate(_ context.Context) (migrateResult, error) {
	cfg, err := config.Load("")
	if err != nil {
		return migrateResult{}, fmt.Errorf("load config: %w", err)
	}

	migrator, err := persistence.NewMigrator(toPersistenceConfig(cfg))
	if err != nil {
		return migrateResult{}, fmt.Errorf("create migrator: %w", err)
	}

	defer func() {
		if err := migrator.Close(); err != nil {
			log.Printf("failed to close migrator: %v", err)
		}
	}()

	if err := migrator.Up(); err != nil {
		return migrateResult{}, err
	}

	version, dirty, err := migrator.Version()
	if errors.Is(err, persistence.ErrNoVersion) {
		return migrateResult{Message: "no migration applied"}, nil
	}

	if err != nil {
		return migrateResult{}, fmt.Errorf("read schema version: %w", err)
	}

	return migrateResult{Version: version, Dirty: dirty, Message: "migrations applied"}, nil
}

// runCLI parses the command line and dispatches one subcommand.
func runCLI() error {
	configPath := flag.String("config", "", "path to config file")
	confirm := flag.Bool("yes", false, "confirm destructive commands (down)")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	migrator, err := persistence.NewMigrator(toPersistenceConfig(cfg))
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	defer func() {
		if err := migrator.Close(); err != nil {
			log.Printf("failed to close migrator: %v", err)
		}
	}()

	if err := run(migrator, args, *confirm); err != nil {
		return fmt.Errorf("%s: %w", args[0], err)
	}

	return nil
}

// run dispatches a single subcommand. It is separated from main so every exit
// path still runs the deferred Close.
func run(migrator *persistence.Migrator, args []string, confirm bool) error {
	switch cmd := args[0]; cmd {
	case "up":
		if err := migrator.Up(); err != nil {
			return err
		}

		return report(migrator, "migrations applied")

	case "down":
		// Rolling everything back drops tables, so require an explicit opt-in.
		if !confirm {
			return errors.New("down rolls back every migration: re-run with -yes to confirm")
		}

		if err := migrator.Down(); err != nil {
			return err
		}

		return report(migrator, "migrations rolled back")

	case "steps":
		if len(args) < 2 {
			return errors.New("steps requires a step count, e.g. steps -1")
		}

		n, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("parse step count %q: %w", args[1], err)
		}

		if n < 0 && !confirm {
			return errors.New("rolling back requires -yes to confirm")
		}

		if err := migrator.Steps(n); err != nil {
			return err
		}

		return report(migrator, "steps applied")

	case "version":
		return report(migrator, "current schema")

	case "force":
		if len(args) < 2 {
			return errors.New("force requires a version, e.g. force 1")
		}

		version, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("parse version %q: %w", args[1], err)
		}

		if err := migrator.Force(version); err != nil {
			return err
		}

		return report(migrator, "version forced")

	default:
		usage()

		return fmt.Errorf("unknown command %q", cmd)
	}
}

// report prints the resulting schema version so every command ends with the
// operator knowing where the database now stands.
func report(migrator *persistence.Migrator, prefix string) error {
	version, dirty, err := migrator.Version()
	if errors.Is(err, persistence.ErrNoVersion) {
		log.Printf("%s: no migration applied", prefix)

		return nil
	}

	if err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}

	log.Printf("%s: version=%d dirty=%t", prefix, version, dirty)

	return nil
}

// toPersistenceConfig maps the load-time config aggregate onto the Identity BC's
// own persistence config, keeping the BC free of shared config types.
func toPersistenceConfig(cfg *config.AppConfig) persistence.Config {
	return persistence.Config{
		Driver:          cfg.Identity.Database.Driver,
		Host:            cfg.Identity.Database.Host,
		Port:            cfg.Identity.Database.Port,
		User:            cfg.Identity.Database.User,
		Password:        cfg.Identity.Database.Password,
		Name:            cfg.Identity.Database.Name,
		SSLMode:         cfg.Identity.Database.SSLMode,
		MaxOpenConns:    cfg.Identity.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Identity.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Identity.Database.ConnMaxLifetime,
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: migrate [-config path] [-yes] <command> [args]

Commands:
  up             apply every pending migration
  down           roll back every applied migration (requires -yes)
  steps <n>      move n migrations forward (n > 0) or backward (n < 0)
  version        print the current schema version
  force <v>      pin the version and clear the dirty flag, without running SQL

Flags:
`)
	flag.PrintDefaults()
}
