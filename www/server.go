package www

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/solar3s/goregen/regenbox"
	"html/template"
	"log"
	"net/http"
	"time"
)

type Server struct {
	ListenAddr string
	Regenbox   *regenbox.RegenBox
	Verbose    bool
	Debug      bool

	wsUpgrader *websocket.Upgrader
	tplFuncs   template.FuncMap
}

func NewServer() *Server {
	return &Server{
		ListenAddr: "localhost:8080",
	}
}

type RegenboxData struct {
	ListenAddr  string
	State       string
	ChargeState string
	Voltage     string
	Config      regenbox.Config
}

func (s *Server) WsSnapshot(w http.ResponseWriter, r *http.Request) {
	conn, err := s.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error subscribing to websocket:", err)
		http.Error(w, "error subscribing to websocket", 500)
		return
	}

	log.Printf("websocket - subscription from %s", conn.RemoteAddr())
	go func(conn *websocket.Conn, s *Server) {
		var err error
		for {
			<-time.After(time.Second * 2)
			err = conn.WriteJSON(s.Regenbox.Snapshot())
			if err != nil {
				log.Printf("websocket - lost connection to %s", conn.RemoteAddr())
				conn.Close()
				return
			}
		}
	}(conn, s)
}

// Config POST: s.Regenbox.SetConfig() (json encoded),
//              Regenbox's must be stopped first
//         GET: gets current s.Regenbox.Config()
func (s *Server) Config(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var cfg regenbox.Config
		err := json.NewDecoder(r.Body).Decode(&cfg)
		if err != nil {
			log.Println("error decoding json:", err)
			http.Error(w, "couldn't decode provided json", http.StatusUnprocessableEntity)
			return
		}

		if !s.Regenbox.Stopped() {
			http.Error(w, "regenbox must be stopped first", http.StatusNotAcceptable)
			return
		}
		err = s.Regenbox.SetConfig(&cfg)
		if err != nil {
			log.Println("error setting config:", err)
			http.Error(w, "error setting config (internal)", http.StatusInternalServerError)
			return
		}
		break
	case http.MethodGet:
		break
	default:
		http.Error(w, fmt.Sprintf("unexpected http-method (%s)", r.Method), http.StatusMethodNotAllowed)
		return
	}

	// encode regenbox config regardless of http method
	w.WriteHeader(200)
	_ = json.NewEncoder(w).Encode(s.Regenbox.Config())
	return
}

// Snapshot encodes snapshot as json to w.
func (s *Server) Snapshot(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(s.Regenbox.Snapshot())
}

func (s *Server) Home(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI != "/" {
		http.NotFound(w, r)
		return
	}

	name := "html/home.html"
	asset, err := Asset(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't load asset: %s", err), http.StatusInternalServerError)
		return
	}

	t, err := template.New(name).Funcs(s.tplFuncs).Parse(string(asset))
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing %s template: %s", name, err), http.StatusInternalServerError)
		return
	}

	state := s.Regenbox.State()
	var tplData = RegenboxData{
		ListenAddr:  s.ListenAddr,
		State:       state.String(),
		ChargeState: "-",
		Voltage:     "-",
		Config:      regenbox.Config{},
	}

	if s.Regenbox != nil {
		i, err := s.Regenbox.ReadVoltage()
		if err == nil {
			tplData.Voltage = fmt.Sprintf("%dmV", i)
			tplData.ChargeState = s.Regenbox.ChargeState().String()
		}
		tplData.Config = s.Regenbox.Config()
	}

	err = t.ExecuteTemplate(w, name, tplData)
	if err != nil {
		http.Error(w, fmt.Sprintf("error executing %s template: %s", name, err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) Start() {
	s.wsUpgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	s.tplFuncs = template.FuncMap{
		"js":   s.RenderJs,
		"css":  s.RenderCss,
		"html": s.RenderHtml,
	}

	go func() {
		watcher := regenbox.NewWatcher(s.Regenbox, regenbox.DefaultWatcherConfig)
		watcher.WatchConn()
	}()
	go func() {
		http.Handle("/subscribe/snapshot", Logger(http.HandlerFunc(s.WsSnapshot), "ws-snapshot", s.Verbose))
		http.Handle("/config", Logger(http.HandlerFunc(s.Config), "config", s.Verbose))
		http.Handle("/snapshot", Logger(http.HandlerFunc(s.Snapshot), "snapshot", s.Verbose))
		http.Handle("/favicon.ico", http.HandlerFunc(NilHandler))
		http.Handle("/", Logger(http.HandlerFunc(s.Home), "www", s.Verbose))
		log.Printf("listening on %s...", s.ListenAddr)
		if err := http.ListenAndServe(s.ListenAddr, nil); err != nil {
			log.Fatal("http.ListenAndServer:", err)
		}
	}()
}
