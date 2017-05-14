package www

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"time"
)

// CustomResponseWriter allows to store current status code of ResponseWriter.
type CustomResponseWriter struct {
	http.ResponseWriter
	Status int
}

func (w *CustomResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *CustomResponseWriter) Write(data []byte) (int, error) {
	return w.ResponseWriter.Write(data)
}

func (w *CustomResponseWriter) WriteHeader(statusCode int) {
	// set w.Status then forward to inner ResposeWriter
	w.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *CustomResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("underlying ResponseWriter is not a Hijacker")
}

func NilHandler(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte{})
	return
}

func WrapCustomRW(wr http.ResponseWriter) http.ResponseWriter {
	if _, ok := wr.(*CustomResponseWriter); !ok {
		return &CustomResponseWriter{
			ResponseWriter: wr,
			Status:         http.StatusOK, // defaults to ok, some servers might not call wr.WriteHeader at all
		}
	}
	return wr
}

func Logger(handler http.Handler, name string, verbose bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t0 := time.Now()
		w = WrapCustomRW(w)
		handler.ServeHTTP(w, r)
		status := w.(*CustomResponseWriter).Status
		if verbose || status != 200 {
			log.Printf("%s- %s %s> (%d) @%s: - agent:%s - %s",
				name, r.Method, r.RequestURI, w.(*CustomResponseWriter).Status,
				r.Header.Get("X-FORWARDED-FOR"), r.Header.Get("USER-AGENT"), time.Since(t0))
		}
	})
}

// renderAsset executes template located at path (assets.go),
// fake-recursively if specified,
// with data if specified,
// and fncs if specified. Returns a value or an error.
func renderAsset(name string,
	fncs template.FuncMap, recur int, data interface{}) (out string, err error) {

	asset, err := Asset(name)
	if err != nil {
		return "", err
	} else if recur == 1 {
		return string(asset), nil
	}

	t, err := template.New(name).Funcs(fncs).Parse(string(asset))
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = t.ExecuteTemplate(buf, name, data)
	if err != nil {
		return "", err
	}
	return string(buf.Bytes()), nil
}

// RenderJs renders un-espaced js asset
func (s *Server) RenderJs(name string, data interface{}) template.JS {
	js, err := renderAsset(name, s.tplFuncs, 0, data)
	if err != nil {
		log.Printf("renderJs %s: %s", name, err)
		return template.JS(fmt.Sprintf("console.error('%s');", err.Error()))
	}
	return template.JS(js)
}

// RenderCss renders un-espaced css asset
func (s *Server) RenderCss(name string, data interface{}) template.CSS {
	css, err := renderAsset(name, s.tplFuncs, 0, data)
	if err != nil {
		log.Printf("renderCss %s: %s", name, err)
		return template.CSS("")
	}
	return template.CSS(css)
}

// RenderHtml renders un-espaced html asset
func (s *Server) RenderHtml(name string, data interface{}) template.HTML {
	html, err := renderAsset(name, s.tplFuncs, 0, data)
	if err != nil {
		return template.HTML(fmt.Sprintf("<span class='error'>'%s'</span>", err.Error()))
	}
	return template.HTML(html)
}
