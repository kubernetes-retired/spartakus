package collector

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/coreos/pkg/health"
	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
	"github.com/thockin/logr"
	"k8s.io/spartakus/pkg/report"
	"k8s.io/spartakus/pkg/version"
)

var (
	CollectorEndpoint = "/api/v1/stats"
	HealthEndpoint    = "/health"
	VersionEndpoint   = "/version"
)

type APIServer struct {
	Log      logr.Logger
	Port     int
	Database Database
}

func (s *APIServer) Run() error {
	logger := logWriter{
		log: s.Log,
	}
	handler := handlers.LoggingHandler(logger, s.newHandler())
	srv := &http.Server{
		Addr:    net.JoinHostPort("", strconv.Itoa(s.Port)),
		Handler: handler,
	}

	s.Log.V(0).Infof("binding to %s", srv.Addr)

	return srv.ListenAndServe()
}

func (s *APIServer) newHandler() http.Handler {
	m := httprouter.New()
	m.Handle("POST", CollectorEndpoint, s.storeRecordHandler())
	m.Handle("GET", HealthEndpoint, s.healthHandler())
	m.Handle("GET", VersionEndpoint, s.versionHandler())
	return m
}

func (s *APIServer) storeRecordHandler() httprouter.Handle {
	handle := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		var rec report.Record
		if err := json.Unmarshal(body, &rec); err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}
		rec.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
		s.logRecord(&rec)

		if err := s.Database.Store(rec); err != nil {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
	return ContentTypeMiddleware(handle, "application/json")
}

func (s *APIServer) logRecord(r *report.Record) {
	j, err := json.Marshal(r)
	if err != nil {
		s.Log.V(9).Infof("failed to decode record: %v", err)
	} else {
		s.Log.V(9).Infof("received record: %s", string(j))
	}
}

func (s *APIServer) healthHandler() httprouter.Handle {
	hc := health.Checker{
		Checks: []health.Checkable{
			&nopHealthCheckable{},
		},
	}

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		hc.ServeHTTP(w, r)
	}
}

type nopHealthCheckable struct{}

func (c *nopHealthCheckable) Healthy() error {
	//FIXME: fill this in
	return nil
}

func (s *APIServer) versionHandler() httprouter.Handle {
	return func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(version.VERSION)); err != nil {
			s.Log.Errorf("failed writing version response: %v", err)
		}
	}
}

type Database interface {
	Store(report.Record) error
}
