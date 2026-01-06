# Go REST API POC

This is a proof-of-concept REST API built in Go, demonstrating a typical backend stack including CRUD, JWT auth, middleware, database access, validations, and more.

---

## Project Requirements

### 1. Environment Setup
- Use **.env** file to store configuration variables:
  - `PORT`
  - `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
  - `JWT_SECRET`
  - `SWAGGER_ENABLED`
- Load environment variables at startup using `os` or a config package.

---

### 2. Database Connection
- Connect to **Postgres** using `pgx` or `sqlc`.
- Implement a **retry mechanism** on startup in case the DB is not ready.
- Include **Docker Compose** for Postgres setup.

---

### 3. Routing and CRUD Endpoints
- Create REST endpoints for a `Person` entity:
  - `POST /person` → Create
  - `GET /person/{id}` → Read
  - `PUT /person/{id}` → Update
  - `DELETE /person/{id}` → Delete
- Implement **query params** for list endpoints (`pageSize`, `pageNumber`, etc).

---

### 4. Middleware
- **Global error handling** middleware.
- **Logging middleware** for request info.
- **Tracing middleware** for request IDs.
- **CORS middleware**.
- **Permission check** middleware (extract `userId` and `role` from JWT).

---

### 5. JWT Authentication
- Secure endpoints with JWT.
- Middleware extracts token, verifies it, and attaches user info to the request context.
- Generate tokens for testing.

---

### 6. Request Validations
- Validate request body JSON structure for all endpoints.
- Validate query parameters and path params.
- Return `400 Bad Request` on invalid input.

---

### 7. Internationalization (i18n)
- Support multiple languages for error messages and responses.
- Default to English if language header is missing.

---

### 8. Swagger Documentation
- Use Swagger annotations to document endpoints.
- Enable Swagger UI at `/swagger/`.

---

### 9. Health Check Endpoint
- `GET /health` → return `200 OK` with basic server info.

---

### 10. Testing
- Unit tests for handlers and services.
- Integration tests hitting API endpoints.

---

### 11. Docker Compose
- Include `docker-compose.yml` to run Postgres locally.
- Configure network so Go app connects to Postgres.

---

### 12. Makefile
- Include common commands:
  - `make run` → run API
  - `make test` → run tests
  - `make docker` → start Postgres with Docker Compose
  - `make swagger` → generate docs

---

## Suggested Implementation Order

1. **Environment & Config Loader** (`.env` file)  
2. **Database Connection + Retry** (Postgres + Docker Compose)  
3. **Routing & Simple Handlers** (no middleware, just CRUD skeleton)  
4. **Middleware** (logging, error handling, tracing)  
5. **Request Validation** (JSON body + query params)  
6. **JWT Auth + Permission Middleware**  
7. **Health Check Endpoint**  
8. **Swagger Documentation**  
9. **Internationalization (i18n)**  
10. **Testing** (unit + integration)  
11. **Makefile** (wrap commands for running, testing, docs, docker)  

---

## Tools & Libraries

| Feature | Suggested Go Library |
|---------|--------------------|
| Router | `chi` |
| JSON Handling | `encoding/json` |
| DB | `pgx` / `sqlc` |
| JWT | `github.com/golang-jwt/jwt/v5` |
| Logging | `log/slog` or `zerolog` |
| Middleware | `chi` custom middleware |
| Swagger | `swaggo/swag` |
| Validation | `go-playground/validator` |
| Environment | `github.com/joho/godotenv` |

---

## Notes
- Keep middleware light and modular.
- Use `context.Context` to carry request-scoped info like user ID and language.
- Error responses should be standardized JSON objects.
- Start small, then layer additional middleware and features.



## Get process id of port lsof -ti :8080