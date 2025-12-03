# Docker Deployment Guide for Render.com

## Repository Configuration

- **Repository:** https://github.com/joan-ouma/feedback_system
- **Branch:** `master`
- **Root Directory:** `/` (repository root)
- **Dockerfile Location:** `./Dockerfile` (in root)

## Render.com Docker Setup

### Step 1: Create Web Service

1. Go to Render Dashboard → "New +" → "Web Service"
2. Connect repository: `https://github.com/joan-ouma/feedback_system`
3. Configure:

   **Basic Settings:**
   - Name: `feedback-sys`
   - Environment: `Docker` ⚠️ **Important: Select Docker, not Go**
   - Region: Choose closest to you
   - Branch: `master` ⚠️ **Use master branch**
   - Root Directory: `/` (leave empty - means repository root)

   **Docker Settings:**
   - Dockerfile Path: `./Dockerfile` (or leave empty - auto-detects)
   - Docker Context: `.` (current directory)

   **Build & Deploy:**
   - Build Command: (Leave empty - Docker handles this)
   - Start Command: (Leave empty - Dockerfile CMD handles this)

### Step 2: Environment Variables

Set these in Render Dashboard → Environment:

```
PORT=8080
MONGODB_URI=mongodb+srv://joan_test:Redwater710@cluster0.kniyd8u.mongodb.net/feedback_sys?retryWrites=true&w=majority
SESSION_SECRET=<generate-with: openssl rand -hex 32>
LLM_API_KEY=<your-openai-or-gemini-api-key>
LLM_API_URL=https://api.openai.com/v1
LLM_MODEL=gpt-4
TEMPLATE_DIR=/root/templates
STATIC_DIR=/root/static
ENABLE_TRACING=false
JAEGER_ENDPOINT=http://localhost:14268/api/traces
```

**Important Docker Paths:**
- `TEMPLATE_DIR=/root/templates` - Templates are copied to `/root/templates` in Docker
- `STATIC_DIR=/root/static` - Static files are copied to `/root/static` in Docker

### Step 3: Deploy

1. Click "Create Web Service"
2. Render will:
   - Detect the Dockerfile
   - Build the Docker image
   - Copy templates and static files to `/root/`
   - Start the container

## Dockerfile Structure

Your Dockerfile uses a multi-stage build:

**Stage 1 (Builder):**
- Uses `golang:1.21-alpine`
- Builds the Go binary: `bin/server`
- Working directory: `/app`

**Stage 2 (Runtime):**
- Uses `alpine:latest`
- Copies binary to: `/root/server`
- Copies templates to: `/root/templates/`
- Copies static files to: `/root/static/`
- Working directory: `/root/`
- Exposes port: `8080`

## Paths in Docker Container

When running in Docker:

| Resource | Path in Container |
|----------|-------------------|
| Binary | `/root/server` |
| Templates | `/root/templates/` |
| Static Files | `/root/static/` |
| Migrations | `/root/migrations/` |
| Working Directory | `/root/` |

## Environment Variables for Docker

The application automatically detects paths, but these env vars ensure correct paths:

- `TEMPLATE_DIR=/root/templates` - Used by the application
- `STATIC_DIR=/root/static` - Used by the application
- `PORT=8080` - Server port (Render sets this automatically)

## Verification Checklist

After deployment:

1. ✅ Check build logs - should show Docker build steps
2. ✅ Check runtime logs - should show "Template directory: /root/templates"
3. ✅ Verify templates load - visit your app URL
4. ✅ Verify static files load - check CSS is loading
5. ✅ Test functionality - sign up, login, etc.

## Troubleshooting Docker Deployment

### Issue: Build fails
- Check Dockerfile syntax
- Verify all files are committed (templates/, static/, etc.)
- Check build logs for specific errors

### Issue: Templates not found
- Verify `TEMPLATE_DIR=/root/templates` is set
- Check that templates/ directory is in git
- Check Dockerfile copies templates correctly

### Issue: Static files 404
- Verify `STATIC_DIR=/root/static` is set
- Check static/ directory is committed
- Verify static file routes in code

### Issue: Port binding error
- Render sets PORT automatically
- Ensure Dockerfile exposes port 8080
- Check environment variable PORT=8080

## Using render.yaml (Optional)

If you want to use `render.yaml` for configuration:

1. Ensure `render.yaml` is in the root directory
2. Set `env: docker` (not `env: go`)
3. Specify `dockerfilePath: ./Dockerfile`
4. Render will use the YAML config

**Note:** The updated `render.yaml` is configured for Docker deployment.

## Quick Deploy Commands

```bash
# 1. Commit all changes
git add .
git commit -m "Configure Docker deployment"

# 2. Push to master branch
git push origin master

# 3. Render will auto-deploy (if auto-deploy is enabled)
# Or manually trigger deployment in Render dashboard
```

## Summary

- **Deployment Type:** Docker
- **Root Directory:** `/` (repository root)
- **Branch:** `master`
- **Dockerfile:** `./Dockerfile`
- **Template Path:** `/root/templates`
- **Static Path:** `/root/static`
- **Port:** `8080`

Your Dockerfile is already correctly configured! Just ensure:
1. Environment is set to "Docker" in Render
2. Branch is set to "master"
3. Environment variables are set (especially TEMPLATE_DIR and STATIC_DIR)

