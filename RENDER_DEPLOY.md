# Deploying to Render

This guide will help you deploy your feedback-sys application to Render.

## Prerequisites

1. A Render account (sign up at https://render.com)
2. Your LLM API key (OpenAI or compatible)
3. Git repository (GitHub, GitLab, or Bitbucket)

## Step 1: Push Your Code to Git

```bash
git init
git add .
git commit -m "Initial commit"
git remote add origin <your-repo-url>
git push -u origin main
```

## Step 2: Create PostgreSQL Database on Render

1. Go to your Render Dashboard
2. Click "New +" → "PostgreSQL"
3. Configure:
   - Name: `feedback-sys-db`
   - Database: `feedback_sys`
   - User: `feedback_user`
   - Plan: Free (or paid if needed)
4. Click "Create Database"
5. **Copy the Internal Database URL** (you'll need this)

## Step 3: Create Web Service on Render

1. Go to Render Dashboard
2. Click "New +" → "Web Service"
3. Connect your Git repository
4. Configure the service:

   **Basic Settings:**
   - Name: `feedback-sys`
   - Environment: `Go`
   - Region: Choose closest to you
   - Branch: `main`
   - Root Directory: `/` (leave empty)

   **Build & Deploy:**
   - Build Command: `go mod download && go build -o bin/server cmd/server/main.go`
   - Start Command: `./bin/server`

   **Environment Variables:**
   Add these environment variables:
   ```
   PORT=8080
   DATABASE_URL=<paste-internal-database-url-from-step-2>
   SESSION_SECRET=<generate-random-string>
   LLM_API_KEY=<your-openai-api-key>
   LLM_API_URL=https://api.openai.com/v1
   LLM_MODEL=gpt-4
   ENABLE_TRACING=false
   JAEGER_ENDPOINT=http://localhost:14268/api/traces
   ```

   **To generate SESSION_SECRET:**
   ```bash
   openssl rand -hex 32
   ```

5. Click "Create Web Service"

## Step 4: Run Database Migrations

After the service is deployed, you need to run migrations:

### Option 1: Using Render Shell (Recommended)

1. Go to your web service in Render Dashboard
2. Click "Shell" tab
3. Run:
   ```bash
   psql $DATABASE_URL < migrations/001_initial_schema.up.sql
   psql $DATABASE_URL < migrations/002_mood_tracking.up.sql
   psql $DATABASE_URL < migrations/003_add_quiz_questions.up.sql
   ```

### Option 2: Using Local psql

1. Get the External Database URL from Render Dashboard
2. Run locally:
   ```bash
   psql <external-database-url> < migrations/001_initial_schema.up.sql
   psql <external-database-url> < migrations/002_mood_tracking.up.sql
   psql <external-database-url> < migrations/003_add_quiz_questions.up.sql
   ```

## Step 5: Verify Deployment

1. Wait for the build to complete (usually 2-5 minutes)
2. Check the logs for any errors
3. Visit your service URL (e.g., `https://feedback-sys.onrender.com`)
4. Test the application:
   - Sign up
   - Log in
   - Test mood tracking
   - Take a quiz

## Environment Variables Reference

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `PORT` | Server port | Yes | 8080 |
| `DATABASE_URL` | PostgreSQL connection string | Yes | - |
| `SESSION_SECRET` | Secret for session encryption | Yes | - |
| `LLM_API_KEY` | OpenAI API key | No* | - |
| `LLM_API_URL` | LLM API endpoint | No | https://api.openai.com/v1 |
| `LLM_MODEL` | LLM model name | No | gpt-4 |
| `ENABLE_TRACING` | Enable OpenTelemetry | No | false |
| `JAEGER_ENDPOINT` | Jaeger endpoint | No | http://localhost:14268/api/traces |

*LLM_API_KEY is optional but required for consultation and recommendations features.

## Troubleshooting

### Build Fails
- Check build logs in Render Dashboard
- Ensure `go.mod` is committed
- Verify build command is correct

### Database Connection Errors
- Verify `DATABASE_URL` is set correctly
- Use Internal Database URL (not External) for better performance
- Check database is running and accessible

### Application Crashes
- Check logs in Render Dashboard
- Verify all environment variables are set
- Ensure migrations have been run

### Static Files Not Loading
- Verify `static/` directory is committed
- Check file paths in templates

## Free Tier Limitations

- Services spin down after 15 minutes of inactivity
- First request after spin-down may take 30-60 seconds
- Database has connection limits
- Consider upgrading for production use

## Updating Your Deployment

1. Push changes to your Git repository
2. Render will automatically detect and deploy
3. Monitor logs for any issues

## Custom Domain (Optional)

1. Go to your service settings
2. Click "Custom Domains"
3. Add your domain
4. Update DNS records as instructed

## Support

- Render Docs: https://render.com/docs
- Render Status: https://status.render.com

