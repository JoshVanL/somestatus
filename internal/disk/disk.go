package disk

import (
	"context"
	"fmt"
	"syscall"
	"time"

	"github.com/go-logr/logr"
)

func Run(ctx context.Context, log logr.Logger, s *string, event chan<- struct{}) error {
	log = log.WithName("disk")

	update := func() error {
		var stat syscall.Statfs_t
		if err := syscall.Statfs("/", &stat); err != nil {
			return err
		}

		*s = fmt.Sprintf(" %.2fG",
			float64(stat.Bavail*uint64(stat.Bsize))/(1024*1024*1024))
		return nil
	}

	if err := update(); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Minute * 2):
			if err := update(); err != nil {
				log.Error(err, "failed to get disk status")
				continue
			}
			go func() { event <- struct{}{} }()
		}
	}
}
