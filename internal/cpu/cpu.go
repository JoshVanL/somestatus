package cpu

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/shirou/gopsutil/cpu"
)

func Run(ctx context.Context, log logr.Logger, s *string, event chan<- struct{}) error {
	log = log.WithName("cpu")

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			percent, err := cpu.PercentWithContext(ctx, time.Second, false)
			if err != nil {
				log.Error(err, "failed to get cpu")
				continue
			}

			*s = fmt.Sprintf(" %.2f%% ", percent[0])
			go func() { event <- struct{}{} }()
		}
	}
}
