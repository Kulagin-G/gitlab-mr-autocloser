package healthcheck

import (
	"github.com/sirupsen/logrus"
	"gitlab-mr-autocloser/src/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGoroutineHealthcheckHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/healthz/liveTest", nil)
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.AutoCloserConfig{
		HealthcheckOptions: config.HealthcheckOptions{
			Liveness: config.Liveness{
				GorMaxNum: 10,
				Path:      "/healthz/liveTest",
			},
		},
	}

	log := logrus.New()
	h := NewListener(cfg, log)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.GoroutineHealthcheckHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{'live': true, 'code': 200}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestDnsHealthcheckHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/healthz/readyTest", nil)
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.AutoCloserConfig{
		HealthcheckOptions: config.HealthcheckOptions{
			Readiness: config.Readiness{
				Path:              "/healthz/readyTest",
				ResolveTimeoutSec: 5,
				URLCheck:          "gitlab.com",
			},
		},
	}

	log := logrus.New()
	h := NewListener(cfg, log)

	rr := httptest.NewRecorder()
	http.HandlerFunc(h.DNShealthcheckHandler).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{'ready': true, 'code': 200}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
