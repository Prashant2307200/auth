# Chatty Auth Service - Endpoints

Examples and brief descriptions for common endpoints.

- POST /auth/register
  - Body: { "email": "user@example.com", "password": "secret" }
  - Response: 201 Created

- POST /api/v1/team/invite
  - Body: { "email": "invitee@example.com", "role": 2 }
  - Response: 201 { "invite_token": "..." }

- GET /api/v1/team/members
  - Response: 200 [ { id, business_id, email, role, status } ]

- PATCH /api/v1/team/members/{id}/role
  - Body: { "role": 3 }
  - Response: 200

- DELETE /api/v1/team/members/{id}
  - Response: 200

- GET /health
  - Legacy health handler returning basic status

- GET /health/live
  - Liveness: checks DB ping and returns simple ok

- GET /health/ready
  - Readiness: returns JSON { status, timestamp, services:{ database, redis } }
