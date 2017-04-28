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

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI != "/" {
		http.Redirect(w, r, "/", http.StatusFound)
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

	if s.Regenbox == nil {
		log.Println("attempting to (re)connect to regenbox...")
		s.Regenbox, err = regenbox.NewRegenBox(nil, nil)
		if err != nil {
			log.Println("couldn't connect to regenbox -", err)
			tplData.State = fmt.Sprintf("couldn't connect to regenbox - %s", err)
		}
	}

	if s.Regenbox != nil {
		i, err := s.Regenbox.ReadVoltage()
		if err != nil {
			tplData.Voltage = "error reading voltage"
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

func (s *Server) Start() error {
	http.Handle("/", Logger(s, "www"))
	log.Printf("Listening on %s...", s.ListenAddr)
	if err := http.ListenAndServe(s.ListenAddr, nil); err != nil {
		return err
	}
	return nil
}
