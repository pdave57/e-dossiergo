# e-Dossier Quick Start Guide

This guide details the steps to set up, configure, bootstrap, and run the e-Dossier API developer environment.

---

## 📋 Prerequisites

Ensure you have the following installed:
* **Go**: version 1.23 or higher ([Download Go](https://go.dev/dl/))
* **Docker & Docker Compose**: ([Install Docker](https://docs.docker.com/get-docker/))
* **PostgreSQL** (Optional, if running locally without Docker): version 16 or higher

---

## ⚙️ 1. Environment Configuration

1. Copy the example environment template file into `.env`:
   ```bash
   cp .env.example .env
   ```

2. Open the `.env` file and review the settings. At a minimum, set a secure `JWT_SECRET` (at least 32 characters long):
   ```ini
   # Generate a key: openssl rand -hex 32
   JWT_SECRET=change-me-to-a-very-long-random-secret-key-at-least-32-chars
   ```

---

## 🐳 2. Running the Application with Docker Compose (Recommended)

Docker Compose starts a PostgreSQL container and compiles/runs the Go API container automatically.

1. Build and run the containers in detached mode:
   ```bash
   docker-compose up -d --build
   ```

2. Check the logs to ensure the server starts and migrations run successfully:
   ```bash
   docker-compose logs -f api
   ```

3. Verify the health check endpoint:
   ```bash
   curl http://localhost:8090/health
   ```
   **Expected Response:**
   ```json
   {"status":"ok","service":"e-dossier"}
   ```

4. Stop the services:
   ```bash
   docker-compose down
   ```

---

## 💻 3. Running the Application Locally (Alternative)

If you prefer to run the Go application directly on your local machine:

1. Ensure a PostgreSQL server is running and a database named `edossier` is created.
2. Export the required variables in your terminal shell:
   ```bash
   export DATABASE_URL=postgres://edossier:secret@localhost:5432/edossier?sslmode=disable
   export JWT_SECRET=change-me-to-a-very-long-random-secret-key-at-least-32-chars
   export PORT=8090
   export APP_ENV=development
   ```
3. Run the Go entrypoint:
   ```bash
   go run ./cmd/server
   ```
   *Note: Database schemas are migrated automatically on startup.*

---

## 🔑 4. Bootstrapping the Admin User & RBAC

Because the application is a multi-tenant administrative system, users require a `state_id` to register, and protected routes require RBAC roles. Since the database starts empty, follow these bootstrapping steps:

### Step 4.1: Seed the State and Admin Role in the Database
Connect to your database (e.g. using `docker exec -it edossier_pg psql -U edossier`) and run the following queries to insert the first state and role, and link the pre-seeded permissions to the role:

```sql
-- 1. Insert a default State (Federal Capital Territory)
INSERT INTO states (id, name, code, country)
VALUES ('fct-state-uuid-123456', 'Federal Capital Territory', 'FCT', 'Nigeria')
ON CONFLICT (id) DO NOTHING;

-- 2. Create the State Administrator Role
INSERT INTO roles (id, state_id, name, code, description, is_system)
VALUES ('admin-role-uuid-123456', 'fct-state-uuid-123456', 'State Administrator', 'STATE_ADMIN', 'Full state administrative rights', true)
ON CONFLICT (id) DO NOTHING;

-- 3. Link all default system permissions to the Administrator Role
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'admin-role-uuid-123456', id FROM permissions
ON CONFLICT DO NOTHING;
```

### Step 4.2: Register the First User via REST API
Submit a registration request using the `state_id` seeded above:

```bash
curl -X POST http://localhost:8090/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "state_id": "fct-state-uuid-123456",
    "email": "admin@edossier.gov.ng",
    "password": "Password@123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```
*Note the returned `"id"` value in the response JSON. We will refer to this as `USER_UUID`.*

### Step 4.3: Elevate User to State Administrator
Assign the administrator role to the user directly in the database using their `USER_UUID`:

```sql
-- Associate user with the State Administrator role
INSERT INTO user_roles (user_id, role_id)
VALUES ('USER_UUID_FROM_REGISTRATION_RESPONSE', 'admin-role-uuid-123456');
```

Now, this user is authenticated and authorized to perform all state-wide administrative endpoints.

---

## 📡 5. Testing the API Flow

Here are curl requests to test standard operations.

### 1. Login to get the JWT Access Token
```bash
curl -X POST http://localhost:8090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@edossier.gov.ng",
    "password": "Password@123"
  }'
```
*Extract the `access_token` from the response. Use this token in the header of subsequent requests:*
`Authorization: Bearer <your_access_token>`

### 2. Verify Your Admin Session (`/auth/me`)
```bash
curl -H "Authorization: Bearer <access_token>" http://localhost:8090/api/v1/auth/me
```

### 3. Create an Educational Zone
```bash
curl -X POST http://localhost:8090/api/v1/states/fct-state-uuid-123456/zones \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Zone A",
    "code": "ZONE-A"
  }'
```
*Copy the zone's `"id"` in the response. We will refer to this as `ZONE_UUID`.*

### 4. Create an LGA (Local Government Area)
```bash
curl -X POST http://localhost:8090/api/v1/states/fct-state-uuid-123456/lgas \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "zone_id": "ZONE_UUID",
    "name": "Abuja Municipal",
    "code": "AMAC"
  }'
```
*Copy the LGA's `"id"` in the response. We will refer to this as `LGA_UUID`.*

### 5. Register a School
```bash
curl -X POST http://localhost:8090/api/v1/schools \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "zone_id": "ZONE_UUID",
    "lga_id": "LGA_UUID",
    "name": "Government Secondary School Garki",
    "code": "GSS-GARKI",
    "category": "COMBINED",
    "ownership": "GOVERNMENT",
    "address": "Garki Area 10, Abuja"
  }'
```
