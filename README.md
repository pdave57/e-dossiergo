# e-Dossier API

An enterprise-grade, high-performance **Student Information System (SIS) and School Management REST API** built in Go, designed for state-level and national education boards.

e-Dossier supports multi-level administrative scopes, granular Role-Based Access Control (RBAC), academic lifecycle management, personnel and student records, and a flexible end-of-term grading and report card generation pipeline.

---

## 🚀 Key Features

* **Hierarchical Administrative Scopes**: Scopes education management from the top down: `State ──> Educational Zone ──> Local Government Area (LGA) ──> School`.
* **Flexible RBAC Model**: User authentication via JWT with rotating refresh tokens. Access permissions are checked at the route level based on granular resource action configurations.
* **School Facility & Asset Management**: Track infrastructure (e.g., libraries, laboratories, sport fields, ICT centers) and their conditions (Good, Fair, Poor, Defunct).
* **Personnel & Staff Tracking**: Manage personnel roles (Principal, Teacher, Counselor, Admin) and maintain historical records of inter-school staff transfers.
* **Student Lifecycle & Level Progression**: Track student registration, class arm (sub-level) enrollments, and academic year progressions (promoted, repeated, graduated).
* **Comprehensive Grading Pipeline**:
  * Customizable continuous assessment (CA1, CA2, CA3) and examination weight configurations.
  * Define state-wide or school-specific grading scales (e.g., min/max ranges, remarks, points).
  * Automatically calculate student positions, average scores, and compile publishable report cards.
* **Demographic & Analytics Reporting**: Aggregate system-wide and state-level statistics including gender distribution and teaching personnel counts.
* **Clean Architecture**: Decoupled domain models, application use-cases, database repositories, and HTTP interfaces to ensure scalability and testability.

---

## 🛠️ Technology Stack

* **Programming Language**: Go (1.23+)
* **HTTP Routing**: [go-chi/chi](https://github.com/go-chi/chi) (v5) — lightweight and fast
* **Database**: PostgreSQL 16+ (using standard Go library `database/sql` and driver `lib/pq` for raw SQL efficiency)
* **Authentication**: JWT token maker using [golang-jwt/jwt](https://github.com/golang-jwt/jwt) (v5)
* **Containerization**: Docker & Docker Compose

---

## 📁 Project Directory Structure

```text
e-dossiergo/
├── cmd/
│   └── server/             # Application entrypoint (wires dependencies and starts HTTP server)
├── config/                 # Structs for env variable parsing, default settings, and validation
├── internal/
│   ├── domain/             # Core business models, entities, and rules (zero external dependencies)
│   ├── application/        # Use cases, application services, and Data Transfer Objects (DTOs)
│   ├── infrastructure/     # SQL Database interactions, repository implementations, and migrations
│   └── interfaces/         # HTTP routes, REST handlers, and custom middleware
├── pkg/                    # Shared utility helper libraries
│   ├── apperror/           # Centralized custom error definitions
│   ├── crypto/             # Encryption and password hashing helpers
│   ├── logger/             # Slog-based structured logger
│   ├── pagination/         # Reusable paginated response wrappers
│   ├── token/              # JWT maker and validation tokens
│   └── validator/          # Input schema and payload verification
├── Dockerfile              # Clean scratch multi-stage build image
├── docker-compose.yml      # Orchestrates Go API and PostgreSQL containers
├── go.mod                  # Go module definition
└── .env.example            # Environment variables configuration blueprint
```

---

## ⚡ Quick Start

For detailed configuration instructions and developer guidelines, refer to the documentation:

* **[Quick Start Guide](quickstart.md)** — Setting up database connections, starting the application, and hitting API routes.
* **[Architecture Document](architecture.md)** — Architectural design patterns, database schema structures, and data flows.
