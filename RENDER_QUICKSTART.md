# Quick Start: Deploy to Render

## 🚀 Fast Deployment Steps

### 1. Push Code to GitHub
```bash
git add .
git commit -m "Ready for Render deployment"
git push origin main
```

### 2. Create Render Account
- Go to https://render.com
- Sign up with GitHub

### 3. Create PostgreSQL Database
1. Dashboard → "New +" → "PostgreSQL"
2. Name: `feedback-sys-db`
3. Plan: Free
4. **Copy Internal Database URL**

### 4. Create Web Service
1. Dashboard → "New +" → "Web Service"
2. Connect your GitHub repo
3. Settings:
   - **Name**: `feedback-sys`
   - **Environment**: `Go`
   - **Build Command**: `go mod download && go build -o bin/server cmd/server/main.go`
   - **Start Command**: `./bin/server`
   - **Plan**: Free

### 5. Set Environment Variables
In Web Service → Environment:
```
PORT=8080
DATABASE_URL=<paste-internal-db-url>
SESSION_SECRET=<run: openssl rand -hex 32>
LLM_API_KEY=<your-openai-key>
LLM_API_URL=https://api.openai.com/v1
LLM_MODEL=gpt-4
ENABLE_TRACING=false
```

### 6. Run Migrations
After first deploy, in Render Shell:
```bash
psql $DATABASE_URL < migrations/001_initial_schema.up.sql
psql $DATABASE_URL < migrations/002_mood_tracking.up.sql
psql $DATABASE_URL < migrations/003_add_quiz_questions.up.sql
```

### 7. Done! 🎉
Visit your app URL: `https://feedback-sys.onrender.com`

## 📝 Important Notes

- **Free tier**: Services sleep after 15 min inactivity
- **Database URL**: Use Internal URL (not External) for better performance
- **Migrations**: Must run manually after first deploy
- **Static files**: Ensure `static/` and `templates/` are committed

## 🔧 Troubleshooting

**Build fails?**
- Check logs in Render Dashboard
- Verify `go.mod` is committed

**Database errors?**
- Verify `DATABASE_URL` is correct
- Check migrations ran successfully

**App not loading?**
- Check service logs
- Verify all env vars are set

## 📚 Full Guide
See `RENDER_DEPLOY.md` for detailed instructions.

