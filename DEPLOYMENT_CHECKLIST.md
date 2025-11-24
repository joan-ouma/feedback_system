# Deployment Checklist for Render

## Pre-Deployment ✅

- [ ] Code is committed and pushed to Git repository
- [ ] All environment variables documented
- [ ] Database migrations are ready
- [ ] Static files and templates are committed

## Render Setup ✅

- [ ] Created Render account
- [ ] Connected GitHub/GitLab repository
- [ ] Created PostgreSQL database
- [ ] Copied Internal Database URL
- [ ] Created Web Service
- [ ] Set all environment variables:
  - [ ] PORT (auto-set by Render, but verify)
  - [ ] DATABASE_URL (Internal URL from Render)
  - [ ] SESSION_SECRET (generated)
  - [ ] LLM_API_KEY (your OpenAI key)
  - [ ] LLM_API_URL
  - [ ] LLM_MODEL
  - [ ] ENABLE_TRACING (set to false)

## Post-Deployment ✅

- [ ] Build completed successfully
- [ ] Service is running
- [ ] Ran database migrations:
  - [ ] 001_initial_schema.up.sql
  - [ ] 002_mood_tracking.up.sql
  - [ ] 003_add_quiz_questions.up.sql
- [ ] Tested application:
  - [ ] Sign up works
  - [ ] Login works
  - [ ] Dashboard loads
  - [ ] Mood tracking works
  - [ ] Quizzes work
  - [ ] Consultation works (if LLM_API_KEY set)

## Environment Variables Template

Copy this and fill in values in Render Dashboard:

```
PORT=8080
DATABASE_URL=<internal-database-url-from-render>
SESSION_SECRET=<generate-with-openssl-rand-hex-32>
LLM_API_KEY=<your-openai-api-key>
LLM_API_URL=https://api.openai.com/v1
LLM_MODEL=gpt-4
ENABLE_TRACING=false
JAEGER_ENDPOINT=http://localhost:14268/api/traces
```

## Quick Commands

**Generate SESSION_SECRET:**
```bash
openssl rand -hex 32
```

**Run migrations in Render Shell:**
```bash
psql $DATABASE_URL < migrations/001_initial_schema.up.sql
psql $DATABASE_URL < migrations/002_mood_tracking.up.sql
psql $DATABASE_URL < migrations/003_add_quiz_questions.up.sql
```

**Or use the script:**
```bash
bash scripts/run-migrations.sh
```

## Troubleshooting

- Check Render logs for errors
- Verify DATABASE_URL format
- Ensure migrations ran successfully
- Check environment variables are set correctly
