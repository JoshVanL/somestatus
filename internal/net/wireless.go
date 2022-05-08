package net

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

const (
	wirelessFilePath = "/proc/net/wireless"
)

func RunWireless(ctx context.Context, log logr.Logger, s *string, event chan<- struct{}) error {
	log = log.WithName("wireless")

	update := func() error {
		w, err := readWireless()
		if err != nil {
			return err
		}

		if len(w) > 0 {
			*s = fmt.Sprintf(" %s%% ", w)
		} else {
			*s = ""
		}

		return nil
	}

	if err := update(); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second * 5):
			if err := update(); err != nil {
				log.Error(err, "failed to get wireless info")
				continue
			}
			go func() { event <- struct{}{} }()
		}
	}
}

func readWireless() (string, error) {
	b, err := os.ReadFile(wirelessFilePath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(b), "\n")
	if len(lines) < 3 {
		return "", nil
	}

	fields := strings.Fields(lines[2])
	return strings.TrimSuffix(fields[2], "."), nil
}
