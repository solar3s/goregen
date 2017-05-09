package regenbox

import (
	"log"
	"time"
)

type Watcher struct {
	rbox   *RegenBox
	cfg    *WatcherConfig
	stopCh chan struct{}
}

type WatcherConfig struct {
	ConnPollRate time.Duration
}

var DefaultWatcherConfig = &WatcherConfig{
	ConnPollRate: time.Second,
}

func NewWatcher(box *RegenBox, cfg *WatcherConfig) *Watcher {
	if cfg == nil {
		cfg = DefaultWatcherConfig
	}
	return &Watcher{
		rbox: box,
		cfg:  cfg,
	}
}

func (w *Watcher) Stop() {
	if w.stopCh == nil {
		return
	}
	log.Println("stopping conn watcher")
	close(w.stopCh)
}

func (w *Watcher) WatchConn() {
	log.Printf("starting conn watcher (poll rate: %s)", w.cfg.ConnPollRate)
	w.stopCh = make(chan struct{})
	var (
		st  State = w.rbox.State()
		err error
	)
	for {
		select {
		case <-time.After(w.cfg.ConnPollRate):
		case <-w.stopCh:
			w.stopCh = nil
			return
		}

		w.rbox.Lock()
		err = w.rbox.ping()
		if err != nil && st == Connected {
			log.Printf("closing serial connection to \"%s\": %s", w.rbox.Conn.path, err)
			w.rbox.Conn.Close()
		}
		st = w.rbox.State()

		switch st {
		case Connected:
		// pass
		default:
			conn, err := FindSerial(nil)
			if err != nil {
				// high-verbosity log
				break
			}
			w.rbox.Conn = conn
			w.rbox.state = Connected
			st = Connected
		}
		w.rbox.Unlock()
	}
}
