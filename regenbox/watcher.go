package regenbox

import (
	"github.com/rkjdid/util"
	"log"
	"sync"
	"time"
)

type Watcher struct {
	rbox   *RegenBox
	cfg    *WatcherConfig
	stopCh chan struct{}
	wg     sync.WaitGroup
}

type WatcherConfig struct {
	ConnPollRate util.Duration
}

var DefaultWatcherConfig = WatcherConfig{
	ConnPollRate: util.Duration(time.Second),
}

func NewWatcher(box *RegenBox, cfg *WatcherConfig) *Watcher {
	if cfg == nil {
		cfg = &DefaultWatcherConfig
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
	w.wg.Wait()
}

func (w *Watcher) WatchConn() {
	w.stopCh = make(chan struct{})
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		var (
			st  State = w.rbox.State()
			err error
		)
		for {
			select {
			case <-time.After(time.Duration(w.cfg.ConnPollRate)):
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

			switch st = w.rbox.State(); st {
			case Connected:
			// pass
			default:
				var conn *SerialConnection
				if w.rbox.Conn != nil && w.rbox.Conn.locked {
					port, cfg, err := OpenPortName(w.rbox.Conn.path)
					if port == nil || err != nil {
						// high-verbosity log
						break
					}
					conn = NewSerial(port, cfg, w.rbox.Conn.path, true)
				} else {
					conn, err = FindSerial(nil)
					if err != nil {
						// high-verbosity log
						break
					}
				}

				w.rbox.Conn = conn
				w.rbox.state = Connected
				st = Connected

				// Restore current charge state.
				// If that fails here, rb.doCycle will
				// remind the box what state it's in.
				_, err = w.rbox.talk(byte(w.rbox.chargeState))
				if err != nil {
					// high-verbosity log
					break
				}
			}
			w.rbox.Unlock()
		}
	}()
}
