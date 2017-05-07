package www

import (
	"fmt"
	"github.com/solar3s/goregen/regenbox"
	"html/template"
	"log"
	"net/http"
)

type Server struct {
	ListenAddr string
	Regenbox   *regenbox.RegenBox
	Verbose    bool
	Debug      bool
}

func NewServer() *Server {
	return &Server{
		ListenAddr: "localhost:8080",
	}
}

type RegenboxData struct {
	State       string
	ChargeState string
	Voltage     string
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

	t, err := template.New(name).Parse(string(asset))
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing %s template: %s", name, err), http.StatusInternalServerError)
		return
	}

	state := s.Regenbox.State()
	var tplData = RegenboxData{
		State:       state.String(),
		ChargeState: "-",
		Voltage:     "-",
	}

	if s.Regenbox != nil {
		i, err := s.Regenbox.ReadVoltage()
		if err != nil {
			log.Printf("ServeHTTP: error reading voltage: \"%s\"", err)
			tplData.Voltage = fmt.Sprintf("error reading voltage: \"%s\"", err)
		} else {
			tplData.Voltage = fmt.Sprintf("%dmV", i)
		}
		tplData.ChargeState = s.Regenbox.ChargeState().String()
	}

	err = t.ExecuteTemplate(w, name, tplData)
	if err != nil {
		http.Error(w, fmt.Sprintf("error executing %s template: %s", name, err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) Start() {
	go func() {
		watcher := regenbox.NewWatcher(s.Regenbox, regenbox.DefaultWatcherConfig)
		watcher.WatchConn()
	}()
	go func() {
		http.Handle("/", Logger(http.HandlerFunc(s.Home), "www", s.Verbose))
		log.Printf("listening on %s...", s.ListenAddr)
		if err := http.ListenAndServe(s.ListenAddr, nil); err != nil {
			log.Fatal("http.ListenAndServer:", err)
		}
	}()
}
