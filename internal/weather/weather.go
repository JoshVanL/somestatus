package weather

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

func Run(ctx context.Context, log logr.Logger, s *string, event chan<- struct{}) error {
	log = log.WithName("weather")

	update := func() {
		resp, err := http.Get("http://wttr.in/London?format=1")
		if err != nil {
			log.Error(err, "failed to get weather report")
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error(err, "failed read weather response")
			return
		}

		*s = strings.TrimSpace(string(body))
		go func() { event <- struct{}{} }()
	}

	for {
		update()
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Minute):
		}
	}
}
