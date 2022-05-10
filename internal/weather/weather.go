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
		resp, err := http.Get("https://wttr.in/?format=%c%t%20%w")
		if err != nil {
			log.Error(err, "failed to get weather report")
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error(err, "failed read weather response")
			return
		}

		*s = strings.TrimSpace(strings.ReplaceAll(string(body), "  ", " "))
		go func() { event <- struct{}{} }()
	}

	// Allow for internet to come up
	select {
	case <-ctx.Done():
		return nil
	case <-time.After(time.Second * 10):
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
