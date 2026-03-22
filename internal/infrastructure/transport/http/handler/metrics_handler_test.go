package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsEndpointReturns200(t *testing.T) {
	mux := http.NewServeMux()
	RegisterMetricsHandler(mux)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
}

func TestMetricsResponseContainsCustomMetrics(t *testing.T) {
	mux := http.NewServeMux()
	RegisterMetricsHandler(mux)

	// Increment some metrics so they appear in output
	inviteSent.Inc()
	inviteAccepted.Inc()
	tokenVerificationsTotal.Inc()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "auth_invites_sent_total") {
		t.Fatalf("metrics output missing auth_invites_sent_total")
	}
	if !strings.Contains(body, "auth_invites_accepted_total") {
		t.Fatalf("metrics output missing auth_invites_accepted_total")
	}
	if !strings.Contains(body, "auth_token_verifications_total") {
		t.Fatalf("metrics output missing auth_token_verifications_total")
	}
}

func TestMetricsContentType(t *testing.T) {
	mux := http.NewServeMux()
	RegisterMetricsHandler(mux)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") && !strings.HasPrefix(ct, "application/openmetrics-text") {
		t.Fatalf("unexpected content-type: %s", ct)
	}
}
