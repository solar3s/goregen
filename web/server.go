package web

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
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	_ "net/http/pprof"
)

const liveInterval = util.Duration(time.Second * 15)
const livePoints = 2400
const liveLog = "data.log"

type ServerConfig struct {
	ListenAddr        string
	Verbose           bool
	StaticDir         string
	DataDir           string
	WebsocketInterval util.Duration

	version string
}

var DefaultServerConfig = ServerConfig{
	ListenAddr:        "localhost:3636",
	WebsocketInterval: util.Duration(time.Second),
}

type Server struct {
	Config   *Config
	Regenbox *regenbox.RegenBox

	liveDataPath string
	liveData     *util.TimeSeries
	cfgPath      string
	router       *mux.Router
	wsUpgrader   *websocket.Upgrader
	tplFuncs     template.FuncMap
	tplData      TemplateData
	cycleSubs    map[int]chan regenbox.CycleMessage
	subId        int
	sync.Mutex
}

type Link struct {
	Href, Name string
}

var HomeLink = Link{
	Href: "/",
	Name: "Live",
}

var ChartsLink = Link{
	Href: "/charts",
	Name: "Charts",
}

type TemplateData struct {
	*Config
	Link    Link
	Version string
}

// StartServer starts a new http.Server using provided version, RegenBox & Config.
// It either doesn't return or panics (http.Listen)
func StartServer(version string, rbox *regenbox.RegenBox, cfg *Config, cfgPath string) {
	if cfg == nil {
		cfg = &DefaultConfig
	}
	cfg.Web.version = version
	srv := &Server{
		Config:   cfg,
		Regenbox: rbox,
		cfgPath:  cfgPath,
	}
	srv.wsUpgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	srv.tplFuncs = template.FuncMap{
		"js":   srv.RenderJs,
		"css":  srv.RenderCss,
		"html": srv.RenderHtml,
	}
	srv.tplData = TemplateData{
		srv.Config, ChartsLink, version,
	}
	srv.cycleSubs = make(map[int]chan regenbox.CycleMessage)

	srv.liveDataPath = filepath.Join(filepath.Dir(srv.cfgPath), liveLog)

	// load previous live data
	err := util.ReadTomlFile(&srv.liveData, srv.liveDataPath)
	if err != nil {
		srv.liveData = util.NewTimeSeries(livePoints, liveInterval)
	} else {
		// shift start time relative to now
		srv.liveData.ResetStartTime()
		// set max length, which is unexported
		srv.liveData.SetMaxLength(livePoints)
	}

	// start voltage monitoring
	go func() {
		ticker := time.NewTicker(time.Duration(srv.liveData.Interval))
		var sn regenbox.Snapshot
		for range ticker.C {
			sn = srv.Regenbox.Snapshot()

			// skip if box isn't connected
			if sn.State != regenbox.Connected {
				continue
			}
			srv.liveData.Add(sn.Voltage)

			// save to file every 10ticks
			err := util.WriteTomlFile(srv.liveData, srv.liveDataPath)
			if err != nil {
				log.Println("couldn't save live datalog:", err)
			}
		}
	}()

	// router
	verbose := srv.Config.Web.Verbose
	srv.router = mux.NewRouter()

	// pprof handlers
	srv.router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	// shh
	srv.router.Handle("/favicon.ico", http.RedirectHandler("/static/img/icon.png", 302))

	// register endpoints
	srv.router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", Logger(http.HandlerFunc(srv.Static), "static", verbose))).
		Methods("GET", "HEAD")
	srv.router.Handle("/websocket",
		Logger(http.HandlerFunc(srv.Websocket), "ws-snapshot", verbose)).
		Methods("GET", "HEAD")
	srv.router.Handle("/config",
		Logger(http.HandlerFunc(srv.RegenboxConfigHandler), "config", verbose)).
		Methods("GET", "POST", "HEAD")
	srv.router.Handle("/start",
		Logger(http.HandlerFunc(srv.StartRegenbox), "start", verbose)).
		Methods("POST", "HEAD")
	srv.router.Handle("/stop",
		Logger(http.HandlerFunc(srv.StopRegenbox), "stop", verbose)).
		Methods("POST", "HEAD")
	srv.router.Handle("/data",
		Logger(http.HandlerFunc(srv.LiveData), "livedata", verbose)).
		Methods("GET", "HEAD")
	srv.router.Handle("/snapshot",
		Logger(http.HandlerFunc(srv.Snapshot), "snapshot", verbose)).
		Methods("GET", "HEAD")
	srv.router.Handle("/charts",
		Logger(http.HandlerFunc(srv.Charts), "charts", verbose)).
		Methods("GET", "HEAD")
	srv.router.Handle("/",
		Logger(http.HandlerFunc(srv.Home), "web", verbose)).
		Methods("GET", "HEAD")

	// http root handle on gorilla router
	httpServer := &http.Server{
		Handler:      srv.router,
		Addr:         srv.Config.Web.ListenAddr,
		WriteTimeout: 4 * time.Second,
		ReadTimeout:  4 * time.Second,
	}
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal("http.ListenAndServer:", err)
	}
}

// Websocket is the handler to initiate a websocket connection
// that keeps track of regenbox state and live measurements.
func (s *Server) Websocket(w http.ResponseWriter, r *http.Request) {
	var interval = time.Duration(s.Config.Web.WebsocketInterval)
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

	if s.Config.Web.Verbose {
		log.Printf("websocket - subscription from %s (pollrate: %s)", conn.RemoteAddr(), interval)
	}

	go func(conn *websocket.Conn, s *Server) {
		var err error
		// subscribe to live ticker
		liveId, liveCh := s.liveData.Subscribe()
		// start state ticker
		ticker := time.NewTicker(interval)
		// subscribe to cycle ticker
		s.Lock()
		cycleCh := make(chan regenbox.CycleMessage, 10)
		cycleId := s.subId
		s.cycleSubs[cycleId] = cycleCh
		s.subId++
		s.Unlock()

		data := struct {
			Type string
			Data interface{}
		}{"state", s.Regenbox.Snapshot()}
		for {
			// send regenbox state asap
			err = conn.WriteJSON(data)
			if err != nil {
				if s.Config.Web.Verbose {
					log.Printf("websocket - lost connection to %s", conn.RemoteAddr())
				}
				conn.Close()
				ticker.Stop()
				s.liveData.Unsubscribe(liveId)
				s.Lock()
				delete(s.cycleSubs, cycleId)
				s.Unlock()
				return
			}

			select {
			case <-ticker.C:
				// type: regenbox.Snapshot
				data.Data = s.Regenbox.Snapshot()
				data.Type = "state"
			case x := <-liveCh:
				// type: int
				data.Data = x
				data.Type = "ticker"
			case msg := <-cycleCh:
				// type: regenbox.CycleMessage
				data.Data = msg
				data.Type = "cycle"
			}
		}
	}(conn, s)
}

// RegenboxConfigHandler POST: s.Regenbox.SetConfig() (json encoded),
//                             Regenbox's must be stopped first
//                       GET: gets current s.Regenbox.Config()
func (s *Server) RegenboxConfigHandler(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, "regenbox must be stopped first", http.StatusConflict)
			return
		}
		err = s.Regenbox.SetConfig(&cfg)
		if err != nil {
			log.Println("error setting config:", err)
			http.Error(w, "error setting config", http.StatusInternalServerError)
			return
		}
		s.Config.Regenbox = cfg

		// save newly set config
		err = util.WriteTomlFile(s.Config, s.cfgPath)
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

func (s *Server) StartRegenbox(w http.ResponseWriter, r *http.Request) {
	err, snaps, messages := s.Regenbox.Start()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte("regenbox started"))

	f := filepath.Join(
		s.Config.Web.DataDir,
		fmt.Sprintf(
			"%s_%s_%s_%s",
			s.Config.User.BetaId,
			s.Config.Battery.BetaRef,
			s.Config.Regenbox.Mode,
			time.Now().Format("2006-01-02_15h04m05.log")))

	go func(file string, cfg regenbox.Config, user User, batt Battery) {
		datalog := util.NewTimeSeries(0, cfg.Ticker)

		var sn regenbox.Snapshot
		var msg regenbox.CycleMessage
		for {
			select {
			case sn = <-snaps:
				if s.Config.Web.Verbose {
					log.Println(sn)
				}
				// add to chart
				datalog.Add(sn.Voltage)
			case msg = <-messages:
				log.Printf("%s: %s", msg.Type, msg.Status)
				s.Lock()
				// broadcast message
				for i, ch := range s.cycleSubs {
					if len(ch) == cap(ch) {
						log.Printf("killing full chan %d", i)
						delete(s.cycleSubs, i)
					} else {
						ch <- msg
					}
				}
				s.Unlock()

				if msg.Final == true {
					chart := ChartLog{
						User:          user,
						Battery:       batt,
						CycleType:     msg.Type,
						TargetReached: !msg.Erronous,
						TotalDuration: util.Duration(datalog.End.Sub(datalog.Start)),
						Reason:        msg.Status,
						Config:        cfg,
						Measures:      *datalog,
					}
					err := util.WriteTomlFile(chart, file)
					if err == nil {
						log.Printf("Saved chart log: %s", file)
					} else {
						log.Printf("Couldn't save chart log %s: %s", file, err)
						log.Println(chart)
					}
					return
				}
			}
		}
	}(f, s.Config.Regenbox, s.Config.User, s.Config.Battery)
}

func (s *Server) StopRegenbox(w http.ResponseWriter, r *http.Request) {
	s.Regenbox.Stop()
	w.Write([]byte("regenbox stopped"))
}

// LiveData encodes live measurement log as json to w.
func (s *Server) LiveData(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(s.liveData.Padded())
}

// Snapshot encodes snapshot as json to w.
func (s *Server) Snapshot(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(s.Regenbox.Snapshot())
}

// Static server
func (s *Server) Static(w http.ResponseWriter, r *http.Request) {
	var err error
	var tpath = filepath.Join(s.Config.Web.StaticDir, r.URL.Path)

	// from s.Static folder
	if f, err := os.Open(tpath); err == nil {
		defer f.Close()
		w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(r.URL.Path)))
		_, err = io.Copy(w, f)
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
	w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(r.URL.Path)))
	_, err = w.Write(asset)
	if err != nil {
		serr := fmt.Sprintf("w.Write %s: %s", tpath, err)
		log.Println(serr)
		http.Error(w, serr, http.StatusInternalServerError)
	}
	return
}

// Home serves homepage
func (s *Server) Home(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = "html/base.html"
	tplFiles := []string{"html/base.html", "html/home.html"}
	data := s.tplData
	data.Link = ChartsLink
	s.makeTplHandler(tplFiles, data, s.tplFuncs).ServeHTTP(w, r)
}

// Explorer page
func (s *Server) Charts(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = "html/base.html"
	tplFiles := []string{"html/base.html", "html/charts.html"}
	data := s.tplData
	data.Link = HomeLink
	s.makeTplHandler(tplFiles, data, s.tplFuncs).ServeHTTP(w, r)
}

// makeStaticHandler creates a handler that tries to load templates
// files from s.StaticDir first, then from binary Assets. It executes successfully
// loaded templates with provided tplData and tplFuncs.
func (s *Server) makeTplHandler(templates []string, tplData interface{}, tplFuncs template.FuncMap) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		var fsTemplates = make([]string, len(templates))
		var assetsTemplates = make([]string, len(templates))
		for i, v := range templates {
			fsTemplates[i] = filepath.Join(s.Config.Web.StaticDir, v)
			assetsTemplates[i] = path.Join("static", v)
		}
		if tplData == nil {
			tplData = s.tplData
		}
		if tplFuncs == nil {
			tplFuncs = s.tplFuncs
		}

		var tname = filepath.Base(r.URL.Path)

		tpl := template.New(tname).Funcs(s.tplFuncs)
		tpl, err = tpl.ParseFiles(fsTemplates...)
		if err != nil {
			// reset tpl and try loading from assets instead
			tpl = template.New(tname).Funcs(s.tplFuncs)
			for _, v := range assetsTemplates {
				asset, err := Asset(v)
				if err != nil {
					http.NotFound(w, r)
					return
				}
				tpl, err = tpl.Parse(string(asset))
				if err != nil {
					serr := fmt.Sprintf("error parsing %s template: %s", r.URL.Path, err)
					log.Println(serr)
					http.Error(w, serr, http.StatusInternalServerError)
					return
				}
			}
		}
		err = tpl.Execute(w, tplData)
		if err != nil {
			serr := fmt.Sprintf("error executing %s template: %s", r.URL.Path, err)
			log.Println(serr)
			http.Error(w, serr, http.StatusInternalServerError)
			return
		}
		return
	})
}
