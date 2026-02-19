# Client Service

React-based frontend application with minimalist macOS-style design.

## Features

- Login and Registration pages
- Profile page with user information
- Semi-transparent bottom tab bar navigation
- GitLab-style default avatar generation
- File upload for avatar images
- Automatic token refresh
- macOS-style design with light/dark mode support

## Development

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## Environment Variables

- `VITE_API_BASE_URL` - API base URL. Default: `/api` (same-origin, proxied via nginx or Vite dev server)
- Override to `http://localhost:8080` to bypass proxy (direct to api-gateway)

## Request Flow

- **Docker**: Client → nginx → api-gateway → backend services
- **Local dev**: Client → Vite proxy → api-gateway → auth-service

## Docker

```bash
# Full stack (from project root)
docker compose up

# Frontend (client + api-gateway) available at http://localhost:3000
```
