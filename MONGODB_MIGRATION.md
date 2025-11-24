# MongoDB Migration Guide

## Status: In Progress

The application is being migrated from PostgreSQL to MongoDB. Currently:
- ✅ Database connection (MongoDB)
- ✅ User model and repository
- ✅ Auth service updated
- ⏳ Other repositories need migration (feedback, consultation, mood, quiz, quote)

## Quick Setup for MongoDB

### 1. Create MongoDB Atlas Account (Free Tier)
1. Go to https://www.mongodb.com/cloud/atlas
2. Create free cluster (M0)
3. Get connection string

### 2. Update Environment Variables

In Render dashboard, set:
- `MONGODB_URI`: Your MongoDB Atlas connection string
  Example: `mongodb+srv://username:password@cluster.mongodb.net/feedback_sys?retryWrites=true&w=majority`

### 3. No Migrations Needed!
MongoDB creates collections automatically on first use.

## Remaining Work

The following repositories need MongoDB migration:
- `feedback_repository.go`
- `consultation_repository.go`
- `mood_repository.go`
- `quiz_repository.go`
- `quote_repository.go`

These will be updated incrementally.

