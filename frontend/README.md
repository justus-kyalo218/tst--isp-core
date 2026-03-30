## Frontend (React + Vite)

### Run locally

```bash
npm install
npm run dev
```

Open the URL Vite prints in the terminal.

### Split User vs Admin Frontend

Use the same build with different envs to separate public user UI from admin login UI.

User frontend (no login/admin UI visible):

```env
VITE_API_URL=https://api.example.com
VITE_ADMIN_PATH=/admin
VITE_ADMIN_HOST=admin.example.com
```

Admin frontend:

```env
VITE_API_URL=https://api.example.com
VITE_ADMIN_PATH=/admin
VITE_ADMIN_HOST=admin.example.com
```

How it works:

1. If `VITE_ADMIN_HOST` is set, login/admin UI is enabled only on that host.
2. If `VITE_ADMIN_HOST` is empty, login/admin UI uses `VITE_ADMIN_PATH` matching.
2. On normal public URL, users only see purchase/register flows.
