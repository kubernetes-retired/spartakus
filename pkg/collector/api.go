package collector

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/coreos/pkg/capnslog"
	"github.com/coreos/pkg/health"
	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"

	"k8s.io/spartakus/pkg/report"
)

var (
	CollectorEndpoint = "/api/v1/stats"
	HealthEndpoint    = "/health"
	VersionEndpoint   = "/version"
)

type APIServer struct {
	Port     int
	Database database
	Version  string
}

//FIXME: need this?
func (s *APIServer) Name() string {
	return "Spartakus Collector"
}

func (s *APIServer) Run() error {
	logger := &logWriter{
		log:   log,
		level: capnslog.INFO,
	}
	handler := handlers.LoggingHandler(logger, s.newHandler())
	srv := &http.Server{
		Addr:    net.JoinHostPort("", strconv.Itoa(s.Port)),
		Handler: handler,
	}

	log.Infof("%s binding to %s", s.Name(), srv.Addr)

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

		if log.LevelAt(capnslog.DEBUG) {
			logRecord(&rec)
		}

		if err := s.Database.Store(rec); err != nil {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
	return ContentTypeMiddleware(handle, "application/json")
}

func logRecord(r *report.Record) {
	//FIXME: this should be done before unmarshalling?
	j, err := json.Marshal(r)
	if err != nil {
		log.Debugf("failed to decode record: %v", err)
	} else {
		log.Debugf("received record: %s", string(j))
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
	//TODO(bcwaldon): fill this in
	return nil
}

func (s *APIServer) versionHandler() httprouter.Handle {
	return func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(s.Version)); err != nil {
			log.Errorf("failed writing version response: %v", err)
		}
	}
}

type database interface {
	Store(report.Record) error
}
