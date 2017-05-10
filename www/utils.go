package www

import (
	"log"
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
		if verbose {
			log.Printf("%s- %s %s> (%d) @%s: - agent:%s - %s",
				name, r.Method, r.RequestURI, w.(*CustomResponseWriter).Status,
				r.Header.Get("X-FORWARDED-FOR"), r.Header.Get("USER-AGENT"), time.Since(t0))
		}
	})
}
