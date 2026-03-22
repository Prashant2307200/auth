package handler

import (
	"errors"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Exported via the /metrics endpoint.
var (
	inviteSent = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_invites_sent_total",
		Help: "Total number of invites sent",
	})
	inviteAccepted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_invites_accepted_total",
		Help: "Total number of invites accepted",
	})
	invitesRevoked = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_invites_revoked_total",
		Help: "Total number of invites revoked",
	})

	tokenVerificationsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_token_verifications_total",
		Help: "Total number of token verification attempts",
	})
	tokenVerificationDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "auth_token_verification_duration_seconds",
		Help:    "Token verification duration in seconds",
		Buckets: prometheus.DefBuckets,
	})
)

func init() {
	// Register collectors to the default registry. If they are already registered
	// (e.g. during tests or multiple package initializations), ignore the
	// AlreadyRegistered error and continue.
	collectors := []prometheus.Collector{inviteSent, inviteAccepted, invitesRevoked, tokenVerificationsTotal, tokenVerificationDuration}
	for _, c := range collectors {
		if err := prometheus.Register(c); err != nil {
			var are prometheus.AlreadyRegisteredError
			if errors.As(err, &are) {
				// already registered, ignore
				continue
			}
			// any other error is unexpected
			panic(err)
		}
	}
}

// RegisterMetricsHandler mounts the Prometheus metrics handler on the provided mux.
func RegisterMetricsHandler(mux *http.ServeMux) {
	mux.Handle("/metrics", promhttp.Handler())
}
