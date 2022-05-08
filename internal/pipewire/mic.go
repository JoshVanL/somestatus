package pipewire

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/auroralaboratories/pulse"
	"github.com/go-logr/logr"
)

func RunMic(ctx context.Context, log logr.Logger, s *string, event chan<- struct{}) error {
	// The underlying pulse C libary shits the bed when running across threads.
	runtime.LockOSThread()

	log = log.WithName("microphone")

	conn, err := pulse.New("somestatus-mic")
	if err != nil {
		return fmt.Errorf("failed to instantiate pulse connection: %s", err)
	}

	subConn, err := pulse.New("somestatus-mic-watcher")
	if err != nil {
		return err
	}
	ch := subConn.Subscribe(pulse.AllEvent)

	for {
		*s, err = updateMic(conn)
		if err != nil {
			log.Error(err, "")
			continue
		}
		go func() { event <- struct{}{} }()

		select {
		case <-ctx.Done():
			if err := conn.Stop(); err != nil {
				log.Error(err, "failed to stop pulse connection")
			}
			if err := subConn.Stop(); err != nil {
				log.Error(err, "failed to stop pulse subscription connection")
			}
			return nil
		case <-ch:
		}
	}
}

func updateMic(conn *pulse.Conn) (string, error) {
	sinks, err := conn.GetSources()
	if err != nil {
		return "-", err
	}

	for _, s := range sinks {
		if strings.Contains(s.Name, "alsa_input") {
			if s.Muted {
				return "", nil
			}
			return "", nil
		}
	}

	return "-", errors.New("no microphone found")
}
