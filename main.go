package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"

	"github.com/joshvanl/somestatus/internal"
)

const xdgEnv = "XDG_RUNTIME_DIR"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	log := stdr.NewWithOptions(
		log.New(os.Stderr, "", log.LstdFlags),
		stdr.Options{LogCaller: stdr.All},
	)
	if err := run(ctx, log); err != nil {
		log.Error(err, "failed to run somestatus modules")
		os.Exit(1)
	}
}

func run(ctx context.Context, log logr.Logger) error {
	path, err := getPath()
	if err != nil {
		return err
	}

	return internal.Run(ctx, log, path)
}

func getPath() (string, error) {
	xdgDir, ok := os.LookupEnv(xdgEnv)
	if !ok {
		return "", fmt.Errorf("environment variable %q not set", xdgEnv)
	}

	for i := 0; i < 100; i++ {
		path := filepath.Join(xdgDir, fmt.Sprintf("somebar-%d", i))
		_, err := os.Stat(path)
		if errors.Is(err, os.ErrNotExist) {
			return filepath.Join(xdgDir, fmt.Sprintf("somebar-%d", i-1)), nil
		}
		if err != nil {
			return "", err
		}
	}

	return "", errors.New("too many somebar fifo files!")
}
