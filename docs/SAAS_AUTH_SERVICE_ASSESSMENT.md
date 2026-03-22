# SaaS Auth Service — Product & Engineering Assessment

This document implements the **SaaS Auth Service Assessment**: evidence review, scorecard, and prioritized roadmap. It is the canonical place for this review (do not rely on ephemeral plan files).

## Scorecard (0–10) — updated after completion pass

| Dimension | Score | Notes |
|-----------|------:|-------|
| **Code quality** | **7.8** | Layering, tests, CI; team routes fixed and wired; profile delete returns 204. |
| **Engineering maturity** | **7.5** | Example configs, README, compose polish, readiness 503, Dockerfile healthcheck + gRPC port. |
| **Architecture** | **7.5** | REST + gRPC registered; team under `/api/v1/team/*` with `X-Tenant-ID` + `TenantFromHeader`. |
| **Product completeness (SaaS auth)** | **6.5** | Core flows + business + team invites; still no reset/verify/MFA/SSO/device sessions. |
| **Overall** | **7.3** | Production-oriented bootstrap; full IdP features remain P2. |

## Evidence — Strengths

- **Structure**: Entry in [`cmd/main/main.go`](../cmd/main/main.go), business logic in [`internal/usecase/`](../internal/usecase/), HTTP in [`internal/infrastructure/transport/http/`](../internal/infrastructure/transport/http/).
- **Tokens**: RS256 access JWT, refresh tokens in Redis with rotation, HttpOnly cookies (see [`internal/service/token.go`](../internal/service/token.go), [`internal/infrastructure/transport/http/utils/response/response.go`](../internal/infrastructure/transport/http/utils/response/response.go)).
- **gRPC**: `TokenService` and `PublicKeyService` registered; codegen under [`internal/transport/grpc/proto/`](../internal/transport/grpc/proto/). Regenerate with `make proto`.
- **Team API**: [`TeamHandler`](../internal/infrastructure/transport/http/handler/team_handler.go) mounted at `/api/v1/team/*`; clients send **`X-Tenant-ID`** (see [`TenantFromHeader`](../internal/infrastructure/transport/http/middleware/tenant.go)).
- **Config**: [`config/local.yaml.example`](../config/local.yaml.example), [`config/docker.yaml`](../config/docker.yaml); required env vars match [`internal/config/config.go`](../internal/config/config.go) (no unused required `ACCESS_TOKEN_SECRET` / `COOKIE_SECRET`).
- **CI**: [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) — lint, race, coverage floor, integration job.
- **Docs**: [`README.md`](../README.md), [`docs/OPERATIONS.md`](OPERATIONS.md), canonical OpenAPI [`api/openapi.yaml`](../api/openapi.yaml); [`docs/openapi.yaml`](openapi.yaml) marked deprecated in-description.

## Evidence — Risks & Gaps

1. **Tenant model**: JWT claims path in `TenantContext` still does not match cookie auth (user id only in context). Mitigation: **`X-Tenant-ID`** for team APIs; long-term unify claims or DB-loaded tenant after auth.
2. **API contract drift**: Add CI OpenAPI vs router checks (P1).
3. **Duplicate RPC**: [`internal/infrastructure/transport/rpc/handler`](../internal/infrastructure/transport/rpc/handler/) legacy `AuthService` — not wired; consolidate or delete (P1).
4. **SaaS feature gaps**: Password reset, email verification, MFA, OIDC, device sessions (P2).

## Prioritized Roadmap

### P0 — Correctness & reliability

- [x] Register gRPC `TokenService` and `PublicKeyService`.
- [x] Example YAML configs + README + compose alignment.
- [x] Stop requiring unused `ACCESS_TOKEN_SECRET` / `COOKIE_SECRET` in env validation.
- [x] Wire team HTTP routes; OpenAPI team paths; `X-Tenant-ID` middleware.
- [x] Readiness returns **503** when dependencies down.
- [ ] Fully align tenant with JWT/DB (optional header-free path).

### P1 — Product confidence

- OpenAPI ↔ runtime contract tests in CI.
- Normalize error JSON across handlers.
- Expiry + refresh client contract tests/docs.

### P2 — SaaS auth completeness

- Password reset + email verification.
- MFA (TOTP), optional OIDC.
- Session/device list + revoke.
- Broader audit usage.

## Success Targets

| Target | Metric |
|--------|--------|
| Transports | HTTP + gRPC documented and exposed (compose maps 9090). |
| API contract | CI fails on OpenAPI vs router drift. |
| Security posture | Reset, verify, MFA, session revocation (P2). |
| Score | **8.2+** after P1–P2 re-audit. |

## Regenerating protobuf / gRPC code

```bash
make proto
```

## Related docs

- [`README.md`](../README.md)
- [`docs/OPERATIONS.md`](OPERATIONS.md)
- [`docs/CONTRIBUTING_TESTING.md`](CONTRIBUTING_TESTING.md)
- [`api/openapi.yaml`](../api/openapi.yaml)
