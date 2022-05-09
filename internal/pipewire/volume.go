package pipewire

import (
	"context"
	"fmt"
	"math"
	"runtime"

	"github.com/auroralaboratories/pulse"
	"github.com/go-logr/logr"
)

func RunVolume(ctx context.Context, log logr.Logger, s *string, event chan<- struct{}) error {
	// The underlying pulse C libary shits the bed when running across threads.
	runtime.LockOSThread()

	log = log.WithName("volume")

	conn, err := pulse.New("somestatus-volume")
	if err != nil {
		return fmt.Errorf("failed to instantiate pulse connection: %s", err)
	}

	subConn, err := pulse.New("somestatus-volume-watcher")
	if err != nil {
		return err
	}
	ch := subConn.Subscribe(pulse.AllEvent)

	for {
		*s, err = updateVolume(conn)
		if err != nil {
			log.Error(err, "failed to update volume")
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

func updateVolume(conn *pulse.Conn) (string, error) {
	sinks, err := conn.GetSinks()
	if err != nil {
		return "-%", err
	}

	if len(sinks) == 0 {
		return "", nil
	}

	lastSink := sinks[len(sinks)-1]

	if lastSink.Muted {
		return " x", nil
	}

	vol := math.Round(100 * float64(lastSink.CurrentVolumeStep) / (100000 / 1.53))

	icon := ""
	if vol == 0 {
		icon = ""
	} else if vol < 40 {
		icon = ""
	}

	return fmt.Sprintf(" %s %.0f%%", icon, vol), nil
}
