# MongoDB Setup for Render

## ✅ Migration Status

**Working:**
- ✅ Signup/Login (User authentication)
- ✅ Feedback submission

**In Progress:**
- ⏳ Consultation, Mood, Quiz, Quote features (temporarily disabled)

## Setup Instructions

### 1. Create MongoDB Atlas Account
1. Go to https://www.mongodb.com/cloud/atlas
2. Sign up for free tier (M0)
3. Create a cluster
4. Get connection string

### 2. Configure Render

In Render dashboard, set environment variable:
- **Key**: `MONGODB_URI`
- **Value**: Your MongoDB Atlas connection string
  Example: `mongodb+srv://username:password@cluster.mongodb.net/feedback_sys?retryWrites=true&w=majority`

### 3. Deploy

Push to GitHub and Render will auto-deploy. No migrations needed - MongoDB creates collections automatically!

## What's Working

- ✅ Anonymous signup with token
- ✅ Login with token
- ✅ Session management
- ✅ Feedback submission
- ✅ View feedbacks

## Next Steps

To complete the migration, update:
- `consultation_repository.go`
- `mood_repository.go`
- `quiz_repository.go`
- `quote_repository.go`

Then uncomment the routes in `cmd/server/main.go`.

