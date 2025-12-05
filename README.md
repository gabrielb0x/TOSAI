# Auto-ToS

Auto-ToS is an AI-powered Terms-of-Service summarizer that delivers quick, digestible insights into website policies. The project includes a Go backend (Gin) and a React + Vite + Tailwind frontend, with PostgreSQL for storage.

## Project structure
- `backend/`: Go service exposing the REST API and background worker (health endpoint available now).
- `frontend/`: Vite + React + TypeScript + Tailwind client scaffold.
- `infra/`: Docker Compose configuration for local development.
- `PLAN.md`: High-level implementation roadmap.

## Getting started with Docker Compose
1. Copy environment defaults and adjust as needed:
   ```bash
   cp .env.example .env
   ```

2. Start the stack (PostgreSQL, backend, frontend):
   ```bash
   cd infra
   docker-compose up --build
   ```
   - Backend: http://localhost:8080/healthz
   - Frontend: http://localhost:5173

3. Stop the stack:
   ```bash
   docker-compose down
   ```

## Running services individually
### Backend
```bash
cd backend
PORT=8080 go run ./cmd/server
```

### Frontend
```bash
cd frontend
npm install
npm run dev -- --host --port 5173
```

## Next steps
- Add database migrations and models.
- Implement summary endpoints and job queue.
- Connect frontend search and admin interfaces to the API.
