# Auth Service API Reference

OpenAPI spec: `api/openapi.yaml`

Quick start

- Run the service locally: `go run ./cmd/main` (or appropriate start command).
- Open the OpenAPI spec in Swagger UI or validate with `swagger-cli validate api/openapi.yaml`.

Authentication

- Protected endpoints require an Authorization header with a Bearer JWT token:

  Authorization: Bearer <accessToken>

- Tokens: the API returns `accessToken` and `refreshToken` from login/register/refresh endpoints.

Endpoints (high level)

- POST /api/v1/auth/register — create user (rate limited: 5 req/min)
- POST /api/v1/auth/login — authenticate user (rate limited: 5 req/min)
- POST /api/v1/auth/logout — revoke tokens
- GET  /api/v1/auth/refresh — rotate/issue access token using refresh token
- GET  /api/v1/auth/profile — get current user profile (protected)
- PUT  /api/v1/auth/profile — update current user (protected)
- DELETE /api/v1/auth/profile — delete current user (protected)
- GET  /api/v1/auth/public-key — public key used to verify tokens
- GET  /api/v1/health — service health
- Users and Business resources under `/api/v1/users` and `/api/v1/business` (protected)

Error handling

- Validation errors return 400 with schema `ValidationError`:

  {
    "errors": [
      {"field": "email", "message": "must be a valid email"}
    ]
  }

- Authentication errors return 401 with `Error` schema:

  {"error": "unauthorized", "message": "invalid token", "statusCode": 401}

- 404 returns `Error` with statusCode 404 when resource not found.

Rate limiting

- Register and Login are rate-limited to 5 requests per minute per IP. If you exceed limits, the service may return 429 Too Many Requests.

Validation and tools

- Validate the OpenAPI file with:

  swagger-cli validate api/openapi.yaml

- Preview using Redocly or Swagger UI by pointing them at `api/openapi.yaml`.

Examples (curl)

- Register

  curl -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d '{"email":"user@example.com","username":"johndoe","password":"P@ssw0rd"}'

- Login

  curl -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"user@example.com","password":"P@ssw0rd"}'

- Get profile

  curl -H "Authorization: Bearer <accessToken>" http://localhost:8080/api/v1/auth/profile

More

- For full request/response schemas and examples, open `api/openapi.yaml` in your preferred OpenAPI viewer.
