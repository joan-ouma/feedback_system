# MongoDB Connection Configuration

## Your MongoDB Atlas Connection Details

**Connection String:**
```
mongodb+srv://joan_test:Redwater710@cluster0.kniyd8u.mongodb.net/feedback_sys?retryWrites=true&w=majority
```

**Database Name:** `feedback_sys`
**Username:** `joan_test`
**Password:** `Redwater710`
**Cluster:** `cluster0.kniyd8u.mongodb.net`

## For Render Deployment

Set this environment variable in Render dashboard:

**Key:** `MONGODB_URI`  
**Value:** `mongodb+srv://joan_test:Redwater710@cluster0.kniyd8u.mongodb.net/feedback_sys?retryWrites=true&w=majority`

## For Local Development (.env file)

Add to your `.env` file:
```
MONGODB_URI=mongodb+srv://joan_test:Redwater710@cluster0.kniyd8u.mongodb.net/feedback_sys?retryWrites=true&w=majority
```

## MongoDB Compass Setup

1. **Download MongoDB Compass** (if not installed):
   - https://www.mongodb.com/try/download/compass

2. **Connect using connection string:**
   - Open MongoDB Compass
   - Paste: `mongodb+srv://joan_test:Redwater710@cluster0.kniyd8u.mongodb.net/feedback_sys?retryWrites=true&w=majority`
   - Click "Connect"

3. **Verify connection:**
   - You should see the `feedback_sys` database
   - Collections will be created automatically when you use the app:
     - `users` - User accounts
     - `feedbacks` - Feedback submissions

## Security Note

⚠️ **Important:** This file contains sensitive credentials. Do NOT commit it to Git!

The connection string is already configured in the code to read from environment variables, so you're safe.

