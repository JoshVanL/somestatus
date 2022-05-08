package datetime

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
)

func Run(ctx context.Context, _ logr.Logger, s *string, event chan<- struct{}) error {
	var until time.Duration
	until, *s = formatTime()
	event <- struct{}{}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(until):
			until, *s = formatTime()
			go func() { event <- struct{}{} }()
		}
	}
}

func formatTime() (time.Duration, string) {
	now := time.Now()
	return time.Until(
			time.Date(
				now.Year(),
				now.Month(),
				now.Day(),
				now.Hour(),
				now.Minute(),
				0,
				0,
				time.Local,
			).Add(time.Minute),
		),
		fmt.Sprintf(`%s %d %s %d %02d:%02d`,
			now.Format("Mon"),
			now.Day(),
			now.Month().String()[:3],
			now.Year(),
			now.Hour(),
			now.Minute(),
		)
}
