## Local Development

### Backend

The backend is a Go service. From the repo root:

```powershell
cd backend
npm.cmd start
```

This works because `backend/package.json` proxies the npm scripts to `go run ./cmd/api`.

The backend requires MongoDB at `mongodb://localhost:27017` by default. A local `backend/.env` file is included for development.

### MongoDB

If you have Docker Desktop installed, start Mongo with:

```powershell
cd infra\mongo
docker compose up -d
```

### Frontend

```powershell
cd frontend
npm.cmd install
npm.cmd run dev
```

### One-command dev helper

On Windows:

```powershell
.\scripts\dev.ps1
```
