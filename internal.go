package servers

import (
	"net/http"
	"time"

	"github.com/bweston92/healthz/healthz"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var LogInternalRequest bool = false

type (
	internalRouter struct {
		hz      *healthz.Healthz
		metrics http.Handler
	}
)

func (h *internalRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if LogInternalRequest {
		logrus.
			WithField("request_uri", r.RequestURI).
			WithField("request_method", r.Method).
			Info("internal endpoint hit")
	}

	switch r.URL.Path {
	case "/healthz":
		h.hz.ServeHTTP(w, r)
	case "/metrics":
		h.metrics.ServeHTTP(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *Server) runInternalHTTP() <-chan error {
	errC := make(chan error)
	h := &http.Server{
		Addr: s.addr,
		Handler: &internalRouter{
			hz:      s.healthz,
			metrics: promhttp.Handler(),
		},
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       300 * time.Second,
	}

	go func() {
		err := h.ListenAndServe()
		if err != nil {
			errC <- err
		}
	}()

	return (<-chan error)(errC)
}
