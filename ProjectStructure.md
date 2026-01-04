## Project Structure

This project follows a modular, scalable Go application layout. The folder structure is organized to separate concerns clearly and support growth over time.

├── cmd/
│ └── api/
│   └── main.go
├── internal/
│ ├── api/
│ │ ├── middleware/
│ │ ├── router/
│ │ └── server.go
│ ├── config/
│ ├── auth/
│ ├── health/
│ ├── product/
│ └── user/
├── pkg/
│ ├── httpUtils/
│ ├── logger/
│ └── timeUtils/
├── .env
├── .env-example
├── .gitignore
├── README.md
└── (future) docker-compose.yml, Makefile, etc.

### `cmd/`
Holds the entrypoints for the application. Each executable has its own folder.

- `cmd/api/main.go`  
  Main entrypoint for the API application. Bootstraps configuration, logging, routing, and server startup.

---

### `internal/`
Contains core application logic. Code here is private to the module and cannot be imported externally.

#### `internal/api/`
Holds everything related to the HTTP API layer.

- `middleware/`  
  Request/response middlewares such as authentication, logging, recovery, etc.

- `router/`  
  Route definitions and organization.

- `server.go`  
  HTTP server setup and lifecycle management.

#### `internal/config/`
Configuration loading, validation, and environment handling.

#### Feature Modules
Each domain or business entity has its own folder, keeping logic encapsulated:

- `auth/`
- `health/`
- `product/`
- `user/`

Typical contents may include:
- Handlers
- Services
- Repository logic
- DTOs / models

---

### `pkg/`
Shared, reusable utilities. These are generic helpers that are not domain-specific.

Examples include:

- `httpUtils/`  
  HTTP helpers and utilities.

- `logger/`  
  Logging setup and helpers.

- `timeUtils/`  
  Time parsing, formatting, and related utilities.

> Code inside `pkg/` may be importable by other modules if needed.

---

### Root-Level Files

- `.env`  
  Local environment configuration.

- `.env-example`  
  Template showing expected environment variables.

- `.gitignore`  
  Files and paths excluded from version control.

- `README.md`  
  Project documentation.

- *(future additions)*  
  - `docker-compose.yml`  
  - `Makefile`  

These will support containerization and development workflow automation.

---

## Goals of This Structure

- Clear separation of responsibilities  
- Maintainable and scalable layout  
- Private application code inside `internal/`  
- Clean entrypoints under `cmd/`  
- Reusable utilities isolated in `pkg/`

This structure makes it easier to add new features, maintain boundaries, and keep the project organized as it grows.
