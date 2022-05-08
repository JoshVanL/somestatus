package net

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

const (
	devPath     = "/sys/class/net"
	statsRXPath = "statistics/rx_bytes"
	statsTXPath = "statistics/tx_bytes"
)

func RunBandwidth(ctx context.Context, log logr.Logger, s *string, event chan<- struct{}) error {
	log = log.WithName("bandwidth")

	ifaceHandler, err := getNetHandler()
	if err != nil {
		return err
	}

	// received, transmitted
	x, err := getBytesX(log, ifaceHandler)
	if err != nil {
		return err
	}

	*s = " 0KiB 0KiB"

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
			currX, err := getBytesX(log, ifaceHandler)
			if err != nil {
				log.Error(err, "failed to get net usage")
				continue
			}

			*s = fmt.Sprintf("%.1fKiB %.1fKiB", (currX[0]-x[0])/(1024), (currX[1]-x[1])/(1024))
			x[0], x[1] = currX[0], currX[1]
			go func() { event <- struct{}{} }()
		}
	}
}

func getBytesX(log logr.Logger, ifaceHandler *netHandler) ([2]float64, error) {
	var currX [2]float64
	for _, iface := range ifaceHandler.ifaces {
		for i, xf := range []string{
			statsRXPath, statsTXPath,
		} {

			fpath := filepath.Join(devPath, iface.Name, xf)
			b, err := os.ReadFile(fpath)
			if err != nil {
				return currX, err
			}

			x, err := strconv.ParseFloat(strings.TrimSpace(string(b)), 64)
			if err != nil {
				return currX, err
			}

			currX[i] += x
		}
	}

	return currX, nil
}
