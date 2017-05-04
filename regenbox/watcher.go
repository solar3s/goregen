package regenbox

import (
	"log"
	"time"
)

type Watcher struct {
	rbox *RegenBox
	cfg  *WatcherConfig
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

func (w *Watcher) WatchConn() {
	log.Printf("starting conn watcher (poll rate: %s)", w.cfg.ConnPollRate)

	var (
		st  State = Connected
		err error
	)
	for {
		select {
		case <-time.After(w.cfg.ConnPollRate):
		case <-w.rbox.stop:
			log.Println("rbox.stop chan closed, watcher out")
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
			st = Connected
		}
		w.rbox.Unlock()
	}
}
