# Deployment Paths & Configuration Guide

## Repository Structure

Your repository: **https://github.com/joan-ouma/feedback_system**

**Branch:** `master` (you're using this branch)

## Root Folder Structure

```
feedback_system/                    ← ROOT DIRECTORY (for deployment)
├── cmd/
│   └── server/
│       └── main.go                  ← Entry point
├── internal/                        ← Application code
├── templates/                       ← HTML templates
├── static/                          ← CSS, JS, images
│   └── css/
│       └── style.css
├── migrations/                      ← Database migrations
├── Dockerfile                       ← Docker build file
├── render.yaml                      ← Render deployment config
├── go.mod                           ← Go dependencies
└── go.sum
```

## Deployment Configuration

### For Render.com Deployment

#### Option 1: Using render.yaml (Recommended)

**Root Directory:** `/` (empty - means repository root)

**Branch:** `master`

**Build Command:** 
```bash
go mod download && go build -ldflags="-s -w" -o bin/server cmd/server/main.go
```

**Start Command:**
```bash
./bin/server
```

**Environment Variables Needed:**
- `PORT=8080` (Render sets this automatically)
- `MONGODB_URI` (your MongoDB connection string)
- `SESSION_SECRET` (generate with: `openssl rand -hex 32`)
- `LLM_API_KEY` (your OpenAI/Gemini API key)
- `LLM_API_URL` (e.g., `https://api.openai.com/v1` or Gemini URL)
- `LLM_MODEL` (e.g., `gpt-4` or `gemini-pro`)
- `TEMPLATE_DIR=/root/templates` (for Docker) or leave unset for native build
- `STATIC_DIR=/root/static` (for Docker) or leave unset for native build

#### Option 2: Manual Render Setup

1. **Root Directory:** Leave empty (defaults to repository root)
2. **Branch:** `master`
3. **Environment:** `Go`
4. **Build Command:** `go mod download && go build -o bin/server cmd/server/main.go`
5. **Start Command:** `./bin/server`

### For Docker Deployment

**Dockerfile Location:** Root directory (`/Dockerfile`)

**Build Context:** Root directory

**Working Directory in Container:** `/root/`

**Template Path:** `/root/templates` (set via `TEMPLATE_DIR` env var)

**Static Path:** `/root/static` (set via `STATIC_DIR` env var)

## Important Paths

### Application Paths (Relative to Root)

- **Entry Point:** `cmd/server/main.go`
- **Templates:** `templates/*.html`
- **Static Files:** `static/css/style.css`
- **Migrations:** `migrations/*.sql`

### Runtime Paths (Inside Container/Server)

When running in Docker:
- Templates: `/root/templates/`
- Static: `/root/static/`
- Binary: `/root/server`

When running natively (Render native build):
- Templates: `./templates/` (relative to where binary runs)
- Static: `./static/` (relative to where binary runs)
- Binary: `./bin/server`

## Render.com Setup Checklist

### Step 1: Connect Repository
- Repository: `https://github.com/joan-ouma/feedback_system`
- Branch: `master` ⚠️ **Important: Use master, not main**
- Root Directory: `/` (leave empty)

### Step 2: Build Settings
- **Build Command:** `go mod download && go build -ldflags="-s -w" -o bin/server cmd/server/main.go`
- **Start Command:** `./bin/server`
- **Docker:** If using Docker, Render will auto-detect Dockerfile

### Step 3: Environment Variables
Set these in Render Dashboard → Environment:

```
PORT=8080
MONGODB_URI=mongodb+srv://...
SESSION_SECRET=<generate-random-string>
LLM_API_KEY=<your-api-key>
LLM_API_URL=https://api.openai.com/v1
LLM_MODEL=gpt-4
TEMPLATE_DIR=./templates
STATIC_DIR=./static
```

### Step 4: Verify Paths
The application will automatically detect:
- Templates directory (checks env var, then working directory, then runtime caller)
- Static files directory (same logic)

## Common Issues & Solutions

### Issue: Templates not found
**Solution:** Ensure `TEMPLATE_DIR` is set correctly:
- For Docker: `TEMPLATE_DIR=/root/templates`
- For native: `TEMPLATE_DIR=./templates` or leave unset (auto-detects)

### Issue: Static files 404
**Solution:** Check that:
1. `static/` directory is committed to git
2. Static file server path is `/static/` in your routes
3. Files exist in `static/css/` directory

### Issue: Build fails
**Solution:** 
1. Ensure `go.mod` and `go.sum` are committed
2. Check build command matches your structure
3. Verify branch is `master`

### Issue: Wrong branch deployed
**Solution:** In Render Dashboard → Settings → Branch, set to `master`

## Quick Reference

| Setting | Value |
|---------|-------|
| Repository Root | `/` (root of repo) |
| Branch | `master` |
| Entry Point | `cmd/server/main.go` |
| Build Output | `bin/server` |
| Templates | `templates/` |
| Static Files | `static/` |
| Port | `8080` |

## Next Steps

1. ✅ Ensure all changes are committed and pushed to `master` branch
2. ✅ Connect repository to Render
3. ✅ Set branch to `master` in Render settings
4. ✅ Configure environment variables
5. ✅ Deploy and test

## Notes

- Your repository has both `main` and `master` branches
- **Always use `master` branch** for deployment
- The `render.yaml` file is configured for the `main` branch - you may need to update it or use manual setup
- Template path resolution works automatically, but setting `TEMPLATE_DIR` explicitly is recommended for Docker deployments

