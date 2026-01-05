## Project Structure

This project follows a modular, scalable Go application layout following industry best practices. The folder structure is organized to separate concerns clearly and support growth over time.

```
rest_api_poc/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── domain/              ← Business domain logic
│   │   ├── product/
│   │   ├── user/
│   │   ├── auth/
│   │   └── health/
│   │
│   ├── infra/               ← Infrastructure layer
│   │   ├── config/
│   │   ├── db/
│   │   ├── router/
│   │   ├── middleware/
│   │   └── server.go
│   │
│   ├── di/                  ← Dependency injection
│   │   └── container.go
│   │
│   └── shared/              ← Shared utilities
│       ├── logger/
│       ├── httpUtils/
│       └── timeUtils/
│
├── assets/
│   └── migrations/
│       └── 001_init.sql
│
├── go.mod
├── go.sum
├── README.md
└── ProjectStructure.md
```

---

## Directory Structure

### `cmd/`
Holds the entrypoints for the application. Each executable has its own folder.

- **`cmd/api/main.go`**  
  Main entrypoint for the API application. Bootstraps configuration, database connection, dependency injection container, and server startup. Handles graceful shutdown.

---

### `internal/`
Contains core application logic. Code here is private to the module and cannot be imported externally.

#### `internal/domain/`
Business domain logic organized by feature/entity. Each domain module is self-contained with its own handlers, services, repositories, and models.

**Structure per domain module:**
- `handler.go` - HTTP request handlers (extract context, validate, call service)
- `service.go` - Business logic layer (implements Service interface)
- `repository.go` - Data access layer (implements Repository interface)
- `model.go` - Domain models/entities
- `routes.go` - Route registration for the domain
- `module.go` - Factory function to wire up the module (optional)

**Current domains:**
- `product/` - Product CRUD operations
- `user/` - User management (to be implemented)
- `auth/` - Authentication (to be implemented)
- `health/` - Health check endpoint

**Pattern:**
```
Handler → Service → Repository → Database
  ↓         ↓          ↓
Context flows through all layers for proper cancellation and timeout handling
```

#### `internal/infra/`
Infrastructure layer - handles external concerns and technical details.

- **`config/`**  
  Configuration loading, validation, and environment variable handling. Loads from `.env` file or system environment.

- **`db/`**  
  Database connection management:
  - `connection.go` - DB interface, connection pool configuration, retry mechanism
  - `setup.go` - Database initialization with retry and graceful shutdown

- **`router/`**  
  HTTP route definitions and organization. Uses chi router, registers all domain routes.

- **`middleware/`**  
  HTTP middleware for cross-cutting concerns:
  - Error handling
  - Logging
  - Tracing
  - CORS
  - Authentication
  - (To be implemented)

- **`server.go`**  
  HTTP server setup, lifecycle management, and graceful shutdown.

#### `internal/di/`
Dependency injection container. Manages all application dependencies and provides factory methods for service modules.

- **`container.go`**  
  Simple dependency injection container. Holds core dependencies (DB, Config) and provides factory methods for domain handlers. Uses manual wiring - simple and explicit, perfect for small to medium applications.

#### `internal/shared/`
Shared utilities used across the application. These are generic helpers that are not domain-specific.

- **`logger/`**  
  Logging setup and helpers. Provides structured logging with different log levels.

- **`httpUtils/`**  
  HTTP-related utilities and helpers.

- **`timeUtils/`**  
  Time parsing, formatting, and related utilities.

---

### `assets/`
Static assets and migration files.

- **`migrations/`**  
  Database migration files (SQL scripts) for schema versioning.

---

## Architecture Patterns

### Dependency Flow
```
main.go
  ↓
di.NewContainer(db, config)
  ↓
infra/server.StartServer(container)
  ↓
infra/router.SetupRouter(container)
  ↓
domain/product.RegisterRoutes(router, container.ProductHandler)
```

### Request Flow
```
HTTP Request
  ↓
Router → Handler (extract context, validate)
  ↓
Service (business logic, context flows through)
  ↓
Repository (data access, context flows through)
  ↓
Database (pgxpool with context)
```

### Context Propagation
Context flows through all layers for:
- Request cancellation (client disconnects)
- Timeout handling
- Request tracing
- Graceful shutdown

---

## Design Principles

1. **Separation of Concerns**
   - Domain logic (`internal/domain/`) is separate from infrastructure (`internal/infra/`)
   - Business logic doesn't depend on HTTP or database details

2. **Dependency Injection**
   - All dependencies are injected, not created
   - Container manages dependencies centrally
   - Easy to test (can mock dependencies)

3. **Interface-Based Design**
   - Services and repositories use interfaces
   - Easy to swap implementations
   - Better testability

4. **Context Propagation**
   - Context flows through all layers
   - Proper cancellation and timeout handling
   - Production-ready error handling

5. **Scalability**
   - Easy to add new domains (just add folder)
   - Container pattern scales to 10-15 services
   - Can migrate to Wire for larger apps

---

## Adding a New Domain

To add a new domain (e.g., `order`):

1. Create `internal/domain/order/` folder
2. Add files: `handler.go`, `service.go`, `repository.go`, `model.go`, `routes.go`
3. Update `internal/di/container.go`:
   - Add `OrderHandler *order.Handler` to Container struct
   - Add `OrderHandler: order.NewModule(database)` to NewContainer
4. Update `internal/infra/router/router.go`:
   - Add `order.RegisterRoutes(r, container.OrderHandler)`

That's it! The pattern is consistent and easy to follow.

---

## Future Enhancements

- [ ] Middleware implementation (error handling, logging, tracing, CORS)
- [ ] Request validation
- [ ] JWT authentication
- [ ] Swagger documentation
- [ ] Internationalization (i18n)
- [ ] Unit and integration tests
- [ ] Docker Compose setup
- [ ] Makefile for common commands

---

## Goals of This Structure

- ✅ Clear separation of responsibilities
- ✅ Maintainable and scalable layout
- ✅ Private application code inside `internal/`
- ✅ Clean entrypoints under `cmd/`
- ✅ Reusable utilities in `internal/shared/`
- ✅ Production-ready patterns and practices
- ✅ Easy to understand and navigate
- ✅ Follows Go community best practices

This structure makes it easier to add new features, maintain boundaries, and keep the project organized as it grows.
