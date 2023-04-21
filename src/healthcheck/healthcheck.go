package healthcheck

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"gitlab-mr-autocloser/src/config"
	"io"
	"log"
	"net"
	"net/http"
	"runtime"
	"time"
)

type HealthCheck interface {
	StartHandlers()
	GoroutineHealthcheckHandler(w http.ResponseWriter, r *http.Request)
	DnsHealthcheckHandler(w http.ResponseWriter, r *http.Request)
}

type healthCheck struct {
	cfg *config.AutoCloserConfig
	log *logrus.Logger
}

func NewListener(cfg *config.AutoCloserConfig, log *logrus.Logger) HealthCheck {
	h := healthCheck{
		cfg: cfg,
		log: log,
	}
	return &h
}

func (h *healthCheck) StartHandlers() {
	addr := fmt.Sprintf("%s:%d", h.cfg.HealthcheckOptions.Host, h.cfg.HealthcheckOptions.Port)
	server := &http.Server{Addr: addr}
	h.log.Infof("Listening healthz on %s", addr)

	http.HandleFunc(h.cfg.HealthcheckOptions.Liveness.Path, h.GoroutineHealthcheckHandler)
	http.HandleFunc(h.cfg.HealthcheckOptions.Readiness.Path, h.DnsHealthcheckHandler)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()
}

// GoroutineHealthcheckHandler returns a 200 OK if there are more than threshold goroutines running
func (h *healthCheck) GoroutineHealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	g := goroutineCount()
	w.Header().Set("Content-Type", "application/json")
	if g > h.cfg.HealthcheckOptions.Liveness.GorMaxNum {
		h.log.Errorf("Liveness failed: goroitines count %d more than gorMaxNum %d!", g, h.cfg.HealthcheckOptions.Liveness.GorMaxNum)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, fmt.Sprintf("{'live': false, 'code': %d}", http.StatusServiceUnavailable))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, fmt.Sprintf("{'live': true, 'code': %d}", http.StatusOK))
}

// DnsHealthcheckHandler resolves the DNS name and returns a 200 OK if successful
func (h *healthCheck) DnsHealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	resolver := net.Resolver{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(h.cfg.HealthcheckOptions.Readiness.ResolveTimeoutSec)*time.Second)
	defer cancel()
	w.Header().Set("Content-Type", "application/json")
	_, err := resolver.LookupHost(ctx, h.cfg.HealthcheckOptions.Readiness.UrlCheck)
	if err != nil {
		h.log.Errorf("Readiness failed: %s was not resolved in %d sec!", h.cfg.HealthcheckOptions.Readiness.UrlCheck, h.cfg.HealthcheckOptions.Readiness.ResolveTimeoutSec)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, fmt.Sprintf("{'ready': false, 'code': %d}", http.StatusServiceUnavailable))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, fmt.Sprintf("{'ready': true, 'code': %d}", http.StatusOK))
}

// goroutineCount returns the number of running goroutines
func goroutineCount() int {
	return runtime.NumGoroutine()
}
