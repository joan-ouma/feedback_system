# URGENT: Fix MongoDB Connection in Render

## Problem
The app is still connecting to `localhost:27017` instead of MongoDB Atlas.

## Root Cause
Render is **NOT** reading the `MONGODB_URI` from `render.yaml`. You must set it manually in the dashboard.

## IMMEDIATE FIX (Do This Now!)

### Step 1: Set Environment Variable in Render Dashboard

1. **Go to Render Dashboard:**
   - https://dashboard.render.com
   - Click on your `feedback-sys` service

2. **Go to Environment Tab:**
   - Click **Environment** in the left sidebar
   - Scroll down to **Environment Variables**

3. **Add MONGODB_URI:**
   - Click **Add Environment Variable** button
   - **Key:** `MONGODB_URI`
   - **Value:** `mongodb+srv://joan_test:Redwater710@cluster0.kniyd8u.mongodb.net/feedback_sys?retryWrites=true&w=majority`
   - Click **Save Changes**

4. **Redeploy:**
   - Render will automatically redeploy
   - OR click **Manual Deploy** → **Deploy latest commit**

### Step 2: Verify MongoDB Atlas Network Access

**CRITICAL:** MongoDB Atlas must allow connections from Render:

1. Go to https://cloud.mongodb.com
2. Click **Network Access** (left sidebar)
3. Click **Add IP Address**
4. Click **Allow Access from Anywhere** (adds `0.0.0.0/0`)
5. Click **Confirm**

### Step 3: Check Logs

After redeploy, check Render logs. You should see:
```
✅ Using MongoDB URI: mongodb+srv://joan_test:***@cluster0.kniyd8u.mongodb.net/feedback_sys
MongoDB connection established
```

If you still see:
```
⚠️  WARNING: MONGODB_URI not set, using default localhost
```

Then the environment variable is still not set correctly.

## Why render.yaml Didn't Work

If you're **not** using Render Blueprint (infrastructure as code), the `render.yaml` file is ignored. You must set environment variables manually in the dashboard.

## Alternative: Use Render Blueprint

If you want to use `render.yaml`:

1. Go to Render Dashboard
2. Click **New** → **Blueprint**
3. Connect your GitHub repo
4. Render will read `render.yaml` and create the service with environment variables

## Quick Checklist

- [ ] `MONGODB_URI` set in Render dashboard (Environment tab)
- [ ] MongoDB Atlas Network Access allows `0.0.0.0/0`
- [ ] Service redeployed after setting environment variable
- [ ] Logs show "Using MongoDB URI" (not localhost warning)

