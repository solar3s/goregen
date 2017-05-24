package www

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rkjdid/util"
	"github.com/solar3s/goregen/regenbox"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

type Server struct {
	ListenAddr string
	Verbose    bool
	Debug      bool
	RboxConfig string
	RootDir    string
	StaticDir  string
	WsInterval time.Duration

	Regenbox *regenbox.RegenBox

	router     *mux.Router
	wsUpgrader *websocket.Upgrader
	tplFuncs   template.FuncMap
}

type RegenboxData struct {
	ListenAddr  string
	State       string
	ChargeState string
	Voltage     string
	Config      regenbox.Config
}

func (s *Server) WsSnapshot(w http.ResponseWriter, r *http.Request) {
	var interval = s.WsInterval
	if v, ok := r.URL.Query()["poll"]; ok {
		if d, err := time.ParseDuration(v[0]); err == nil {
			interval = d
		}
	}
	conn, err := s.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error subscribing to websocket:", err)
		http.Error(w, "error subscribing to websocket", 500)
		return
	}

	if s.Verbose {
		log.Printf("websocket - subscription from %s", conn.RemoteAddr())
	}

	go func(conn *websocket.Conn, s *Server) {
		var err error
		for {
			err = conn.WriteJSON(s.Regenbox.Snapshot())
			if err != nil {
				if s.Verbose {
					log.Printf("websocket - lost connection to %s", conn.RemoteAddr())
				}
				conn.Close()
				return
			}
			<-time.After(interval)
		}
	}(conn, s)
}

// Config POST: s.Regenbox.SetConfig() (json encoded),
//              Regenbox's must be stopped first
//         GET: gets current s.Regenbox.Config()
func (s *Server) Config(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// copy current config, this allows for setting only a subset of the whole config
		var cfg regenbox.Config = s.Regenbox.Config()
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
		// save newly set config
		err = util.WriteTomlFile(cfg, s.RboxConfig)
		if err != nil {
			log.Println("error writing config:", err)
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

// Static server
func (s *Server) Static(w http.ResponseWriter, r *http.Request) {
	var err error
	var tpath = filepath.Join(s.StaticDir, r.URL.Path)

	// from s.Static folder
	if f, err := os.Open(tpath); err == nil {
		defer f.Close()
		_, err := io.Copy(w, f)
		if err != nil {
			serr := fmt.Sprintf("io.Copy %s: %s", tpath, err)
			log.Println(serr)
			http.Error(w, serr, 500)
		}
		return
	}

	// from binary assets
	asset, err := Asset(path.Join("static", r.URL.Path))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_, err = w.Write(asset)
	if err != nil {
		serr := fmt.Sprintf("w.Write %s: %s", tpath, err)
		log.Println(serr)
		http.Error(w, serr, http.StatusInternalServerError)
	}
	return
}

func (s *Server) Home(w http.ResponseWriter, r *http.Request) {
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

	// set path to home template in request
	r.URL.Path = "html/home.html"
	s.makeTplHandler(tplData).ServeHTTP(w, r)
}

// makeStaticHandler creates a handler that tries to load r.URL.Path
// file from s.StaticDir first, then from Assets. It executes successfully
// loaded template with profided tplData.
func (s *Server) makeTplHandler(tplData interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		var tpath = filepath.Join(s.StaticDir, r.URL.Path)
		var tname = filepath.Base(r.URL.Path)

		tpl := template.New(tname).Funcs(s.tplFuncs)
		tpl2, err := tpl.ParseFiles(tpath)
		if err != nil {
			// try loading asset instead
			asset, err := Asset(path.Join("static", r.URL.Path))
			if err != nil {
				http.NotFound(w, r)
				return
			}
			tpl2, err = tpl.Parse(string(asset))
			if err != nil {
				serr := fmt.Sprintf("error parsing %s template: %s", r.URL.Path, err)
				log.Println(serr)
				http.Error(w, serr, http.StatusInternalServerError)
				return
			}
		}

		err = tpl2.ExecuteTemplate(w, tname, tplData)
		if err != nil {
			serr := fmt.Sprintf("error executing %s template: %s", r.URL.Path, err)
			log.Println(serr)
			http.Error(w, serr, http.StatusInternalServerError)
			return
		}
		return
	})
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
	s.router = mux.NewRouter()

	go func() {
		watcher := regenbox.NewWatcher(s.Regenbox, regenbox.DefaultWatcherConfig)
		watcher.WatchConn()
	}()
	go func() {
		s.router.PathPrefix("/static/").Handler(
			http.StripPrefix("/static/", Logger(http.HandlerFunc(s.Static), "static", s.Verbose))).
			Methods("GET")
		s.router.Handle("/subscribe/snapshot",
			Logger(http.HandlerFunc(s.WsSnapshot), "ws-snapshot", s.Verbose)).
			Methods("GET")
		s.router.Handle("/config",
			Logger(http.HandlerFunc(s.Config), "config", s.Verbose)).
			Methods("GET", "POST")
		s.router.Handle("/snapshot",
			Logger(http.HandlerFunc(s.Snapshot), "snapshot", s.Verbose)).
			Methods("GET")
		s.router.Handle("/favicon.ico", http.HandlerFunc(NilHandler))
		s.router.Handle("/",
			Logger(http.HandlerFunc(s.Home), "www", s.Verbose)).
			Methods("GET")

		// http root handle on gorilla router
		srv := &http.Server{
			Handler:      s.router,
			Addr:         s.ListenAddr,
			WriteTimeout: 4 * time.Second,
			ReadTimeout:  4 * time.Second,
		}
		log.Printf("listening on %s...", s.ListenAddr)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("http.ListenAndServer:", err)
		}
	}()
}
