package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/spf13/pflag"
	"libdb.so/catglow"
)

var (
	config  = "catglow.toml"
	verbose = false
)

func init() {
	pflag.StringVarP(&config, "config", "c", config, "configuration file")
	pflag.BoolVarP(&verbose, "verbose", "v", verbose, "verbose output")
}

func main() {
	pflag.Parse()

	logLevel := slog.LevelWarn
	if verbose {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := readConfig()
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// TODO: add a file detector for when /dev/ttyUSB0 is not available, and
	// automatically start the daemon when it is available.

	d, err := catglow.NewDaemon(cfg, slog.Default())
	if err != nil {
		return fmt.Errorf("failed to create daemon: %w", err)
	}

	if err := d.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("daemon failed: %w", err)
	}

	return nil
}

func readConfig() (*catglow.Config, error) {
	f, err := os.Open(config)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	return catglow.ParseConfig(f)
}
