package util

import (
	"context"
	"sync"

	"github.com/go-logr/logr"
	"github.com/howeyc/fsnotify"
)

var (
	watcher *fsnotify.Watcher
	wg      sync.WaitGroup
)

func init() {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
}

func RunWatcher(ctx context.Context, log logr.Logger) {
	log = log.WithName("file_watcher")
	log.Info("running file_watcher")

	<-ctx.Done()
	log.Info("closing file watcher")
	if err := watcher.Close(); err != nil {
		log.Error(err, "error closing file_watcher")
	}
	wg.Wait()
}

func AddWatcher(ctx context.Context, log logr.Logger, path string) (<-chan struct{}, error) {
	log = log.WithName("file_watcher").WithValues("path", path)
	ch := make(chan struct{})
	wg.Add(1)

	if err := watcher.Watch(path); err != nil {
		return nil, err
	}

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-watcher.Event:
				ch <- struct{}{}
			case err := <-watcher.Error:
				log.Error(err, "watching file")
			}
		}
	}()

	return ch, nil
}
