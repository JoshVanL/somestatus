package temp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/shirou/gopsutil/host"
)

func Run(ctx context.Context, log logr.Logger, s *string, event chan<- struct{}) error {
	log = log.WithName("temperature")

	update := func() {
		temps, err := host.SensorsTemperaturesWithContext(ctx)
		if err != nil {
			log.Error(err, "failed to get temperature")
			return
		}

		var total, i float64
		for _, temp := range temps {
			if strings.HasSuffix(temp.SensorKey, "input") {
				total += temp.Temperature
				i++
			}
		}

		*s = fmt.Sprintf("%1.f°C", total/i)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second * 5):
			update()
			go func() { event <- struct{}{} }()
		}
	}
}
