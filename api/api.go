package api // import "github.com/CenturyLinkLabs/dray/api"

import (
	"fmt"
	"net/http"

	"github.com/CenturyLinkLabs/dray/job"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

type handler func(jm job.Manager, r *http.Request, w http.ResponseWriter)

type statusLoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusLoggingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// A Server is the HTTP server which reponds to Dray API requests.
type Server interface {
	Start(port int)
}

type jobServer struct {
	jobManager job.Manager
}

// NewServer returns a new Server instance which will handle Dray service
// requests and defer work to the specified Manager.
func NewServer(jm job.Manager) Server {
	return &jobServer{jobManager: jm}
}

func (s *jobServer) Start(port int) {
	router := s.createRouter()

	log.Infof("Server running on port %d", port)
	portString := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(portString, router))
}

func (s *jobServer) createRouter() *mux.Router {
	router := mux.NewRouter()

	m := map[string]map[string]handler{
		"GET": {
			"/jobs":             listJobs,
			"/jobs/{jobid}":     getJob,
			"/jobs/{jobid}/log": getJobLog,
		},
		"POST": {
			"/jobs": createJob,
		},
		"DELETE": {
			"/jobs/{jobid}": deleteJob,
		},
	}

	for method, routes := range m {
		for route, fct := range routes {

			localMethod := method
			localRoute := route
			localFct := fct
			wrap := func(w http.ResponseWriter, r *http.Request) {
				ww := &statusLoggingResponseWriter{w, 200}

				log.Infof("Started %s %s", r.Method, r.RequestURI)

				if localMethod != "DELETE" {
					w.Header().Set("Content-Type", "application/json")
				}

				localFct(s.jobManager, r, ww)

				log.Infof("Completed %d", ww.statusCode)
			}
			router.Path("/v{version:[0-9.]+}" + localRoute).Methods(localMethod).HandlerFunc(wrap)
			router.Path(localRoute).Methods(localMethod).HandlerFunc(wrap)
		}
	}

	return router
}
