package memory

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

// stats represents memory statistics for linux
type stats struct {
	total, used, buffers, cached, free, available uint64

	memAvailableEnabled bool
}

func Run(ctx context.Context, log logr.Logger, s *string, event chan<- struct{}) error {
	log = log.WithName("memory")

	update := func() error {
		stats, err := collectMemoryStats()
		if err != nil {
			return err
		}
		*s = fmt.Sprintf(" %.1f/%.1fGb ",
			(float64(stats.used) * 1e-6),
			(float64(stats.total) * 1e-6),
		)
		go func() { event <- struct{}{} }()
		return nil
	}

	if err := update(); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
			if err := update(); err != nil {
				log.Error(err, "failed to gather memory")
				continue
			}
			go func() { event <- struct{}{} }()
		}
	}
}

func collectMemoryStats() (*stats, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var memory stats
	memStats := map[string]*uint64{
		"MemTotal":     &memory.total,
		"MemFree":      &memory.free,
		"MemAvailable": &memory.available,
		"Buffers":      &memory.buffers,
		"Cached":       &memory.cached,
	}
	for scanner.Scan() {
		line := scanner.Text()
		i := strings.IndexRune(line, ':')
		if i < 0 {
			continue
		}
		fld := line[:i]
		if ptr := memStats[fld]; ptr != nil {
			val := strings.TrimSpace(strings.TrimRight(line[i+1:], "kB"))
			*ptr, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return nil, err
			}

			if fld == "MemAvailable" {
				memory.memAvailableEnabled = true
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan error for /proc/meminfo: %s", err)
	}

	if memory.memAvailableEnabled {
		memory.used = memory.total - memory.free
	} else {
		memory.used = memory.total - memory.free - memory.buffers - memory.cached
	}

	return &memory, nil
}
