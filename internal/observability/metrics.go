package observability

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var (
	tracer = otel.Tracer("github.com/Prashant2307200/auth-service")
)

type Metrics struct {
	TokenVerificationsTotal   prometheus.Counter
	TokenVerificationDuration prometheus.Histogram
	InvitesSentTotal          prometheus.Counter
	InvitesAcceptedTotal      prometheus.Counter
	InvitesRevokedTotal       prometheus.Counter
}

func NewMetrics() (*Metrics, error) {
	tokenVerifTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_token_verifications_total",
		Help: "Total number of token verification attempts",
	})

	tokenVerifDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "auth_token_verification_duration_seconds",
		Help:    "Token verification duration in seconds",
		Buckets: prometheus.DefBuckets,
	})

	invitesSent := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_invites_sent_total",
		Help: "Total number of invites sent",
	})

	invitesAccepted := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_invites_accepted_total",
		Help: "Total number of invites accepted",
	})

	invitesRevoked := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_invites_revoked_total",
		Help: "Total number of invites revoked",
	})

	reg := prometheus.NewRegistry()
	reg.MustRegister(tokenVerifTotal, tokenVerifDuration, invitesSent, invitesAccepted, invitesRevoked)

	return &Metrics{
		TokenVerificationsTotal:   tokenVerifTotal,
		TokenVerificationDuration: tokenVerifDuration,
		InvitesSentTotal:          invitesSent,
		InvitesAcceptedTotal:      invitesAccepted,
		InvitesRevokedTotal:       invitesRevoked,
	}, nil
}

func StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, func()) {
	spanCtx, span := tracer.Start(ctx, name)
	for _, attr := range attrs {
		span.SetAttributes(attr)
	}
	return spanCtx, func() { span.End() }
}

func RecordLatency(histogram prometheus.Histogram, duration time.Duration) {
	histogram.Observe(duration.Seconds())
}
