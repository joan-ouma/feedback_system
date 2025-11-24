# Fix Render Environment Variable Issue

## Problem
The app is connecting to `localhost:27017` instead of MongoDB Atlas, which means `MONGODB_URI` is not set in Render.

## Solution

### Option 1: Set in Render Dashboard (Recommended)

1. Go to https://dashboard.render.com
2. Select your `feedback-sys` service
3. Click on **Environment** tab
4. Click **Add Environment Variable**
5. Set:
   - **Key:** `MONGODB_URI`
   - **Value:** `mongodb+srv://joan_test:Redwater710@cluster0.kniyd8u.mongodb.net/feedback_sys?retryWrites=true&w=majority`
6. Click **Save Changes**
7. Render will automatically redeploy

### Option 2: Update render.yaml (Alternative)

The `render.yaml` file has been updated with your MongoDB URI. If you're using Render Blueprint:

1. Commit the updated `render.yaml`
2. Push to GitHub
3. Render will use the environment variable from the YAML

## Verify MongoDB Atlas Network Access

**CRITICAL:** Make sure MongoDB Atlas allows connections from Render:

1. Go to https://cloud.mongodb.com
2. Select your cluster (`cluster0`)
3. Click **Network Access** (left sidebar)
4. Click **Add IP Address**
5. Click **Allow Access from Anywhere** (adds 0.0.0.0/0)
6. Click **Confirm**

Without this, Render won't be able to connect even with the correct connection string!

## Test Connection

After setting the environment variable and allowing network access:

1. Check Render logs - should see: `MongoDB connection established`
2. Try accessing your app - signup should work!

## Troubleshooting

**Still seeing localhost error?**
- Verify `MONGODB_URI` is set in Render dashboard (check Environment tab)
- Make sure there are no typos in the connection string
- Check Render logs for the actual connection string being used

**Connection timeout?**
- Verify MongoDB Atlas Network Access allows all IPs (0.0.0.0/0)
- Check that your MongoDB Atlas cluster is running
- Verify username/password are correct

