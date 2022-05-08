package internal

import (
	"context"
	"os"
	"sync"

	"github.com/go-logr/logr"
	"github.com/joshvanl/somestatus/internal/cpu"
	"github.com/joshvanl/somestatus/internal/datetime"
	"github.com/joshvanl/somestatus/internal/disk"
	"github.com/joshvanl/somestatus/internal/memory"
	"github.com/joshvanl/somestatus/internal/net"
	"github.com/joshvanl/somestatus/internal/pipewire"
	"github.com/joshvanl/somestatus/internal/temp"
	"github.com/joshvanl/somestatus/internal/weather"
)

type runModuleFn func(context.Context, logr.Logger, *string, chan<- struct{}) error

var allModules = [][]runModuleFn{
	[]runModuleFn{
		weather.Run,
	},
	[]runModuleFn{
		pipewire.RunMic,
		pipewire.RunVolume,
	},
	[]runModuleFn{
		cpu.Run,
		memory.Run,
		disk.Run,
	},
	[]runModuleFn{
		net.RunWireless,
		net.RunBandwidth,
	},
	[]runModuleFn{
		temp.Run,
	},
	[]runModuleFn{
		datetime.Run,
	},
}

type handler struct {
	path    string
	eventCh chan struct{}
	errCh   chan error
	wg      sync.WaitGroup

	blocks [][]*string
}

func Run(ctx context.Context, log logr.Logger, path string) error {
	log = log.WithName("modules")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	h := &handler{
		path:    path,
		eventCh: make(chan struct{}),
		errCh:   make(chan error),
	}

	for _, mblocks := range allModules {
		var block []*string
		for _, module := range mblocks {
			h.wg.Add(1)
			block = append(block, h.runModule(ctx, log, module))
		}
		h.blocks = append(h.blocks, block)
	}

	log.Info("all modules running ...")

	var err error
	func() {
		for {
			select {
			case err = <-h.errCh:
				log.Error(err, "recived module error, shutting down ...")
				cancel()
				return
			case <-ctx.Done():
				log.Info("shutting down ...")
				return
			case <-h.eventCh:
				if err := h.update(); err != nil {
					log.Error(err, "failed to update status")
				}
			}
		}
	}()

	h.wg.Done()
	return err
}

func (h *handler) update() error {
	var status string
	for i, block := range h.blocks {
		if i != 0 {
			status += " | "
		}
		for _, s := range block {
			status += *s
		}
	}

	f, err := os.OpenFile(h.path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString("status " + status + "\n"); err != nil {
		return err
	}

	return nil
}

func (h *handler) runModule(ctx context.Context, log logr.Logger, module runModuleFn) *string {
	s := new(string)

	go func() {
		defer h.wg.Done()
		if err := module(ctx, log, s, h.eventCh); err != nil {
			h.errCh <- err
		}
	}()

	return s
}
