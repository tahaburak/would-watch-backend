# Agent Guide: Backend Repository

## 🧠 Context
This is the **Golang API** for Would Watch. It follows Clean Architecture principles to ensure testability and separation of concerns.

## 🏗 Structure
- **`cmd/server/`**: Entry point (`main.go`).
- **`internal/`**: Application logic.
  - **`api/`**: HTTP Handlers and Routes (Chi Router).
  - **`middleware/`**: Auth (`AuthMiddleware`), Logging, CORS.
  - **`services/`**: Business logic (Use Cases).
  - **`repository/`**: Database interactions (Supabase/Postgres).
  - **`models/`**: Domain structs.
- **`db/`**: SQL Schema (`schema.sql`) and migrations.

## 🔑 Key Facts for Agents
1.  **Database**: Postgres (via Supabase). Checks `DATABASE_URL` for connection.
2.  **Auth**: Validates Supabase JWTs. Checks `SUPABASE_JWT_SECRET` or JWKS.
3.  **Testing**:
    - **Unit**: `go test ./...`
    - **Integration**: `tests/integration/`
    - **E2E**: `tests/e2e/` (requires running container/DB).

## 🛠 Common Tasks
- **Run Local**: `go run cmd/server/main.go` (Port 8080).
- **Add Feature**:
    1. Define Model.
    2. Create Repository Interface & Implementation.
    3. Create Service.
    4. Create Handler.
    5. Register Route.
