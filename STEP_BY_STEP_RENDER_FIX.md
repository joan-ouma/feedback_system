# Step-by-Step: Fix MongoDB Connection in Render

## The Problem
Your logs show: `⚠️  WARNING: MONGODB_URI not set, using default localhost`

This means Render is **NOT** reading the environment variable. You must set it manually.

## EXACT Steps to Fix (Follow Carefully)

### Step 1: Open Render Dashboard
1. Go to: https://dashboard.render.com
2. Log in if needed

### Step 2: Find Your Service
1. Click **Dashboard** in the top menu
2. Find **feedback-sys** in the list
3. **Click on it** to open the service page

### Step 3: Go to Environment Tab
1. Look at the left sidebar
2. Click **Environment** (it's below "Events" and above "Logs")
3. You should see a section called **Environment Variables**

### Step 4: Add MONGODB_URI
1. Scroll down to **Environment Variables** section
2. Click the **Add Environment Variable** button (usually blue/green button)
3. In the popup/form:
   - **Key field:** Type exactly: `MONGODB_URI`
   - **Value field:** Paste exactly: `mongodb+srv://joan_test:Redwater710@cluster0.kniyd8u.mongodb.net/feedback_sys?retryWrites=true&w=majority`
4. Click **Save** or **Add**

### Step 5: Verify It Was Added
- You should now see `MONGODB_URI` in the environment variables list
- The value should show (might be masked with dots)

### Step 6: Redeploy
1. Render should auto-redeploy (watch the top right for deployment status)
2. OR click **Manual Deploy** → **Deploy latest commit**

### Step 7: Check Logs
1. Go to **Logs** tab (left sidebar)
2. Wait for deployment to finish
3. Look for:
   - ✅ `Using MongoDB URI: mongodb+srv://joan_test:***@cluster0...` (SUCCESS)
   - ❌ `⚠️  WARNING: MONGODB_URI not set` (FAILED - try again)

## If You Still See the Warning

### Check These:
1. **Spelling:** Is it exactly `MONGODB_URI`? (case-sensitive, no spaces)
2. **Value:** Copy the entire connection string (starts with `mongodb+srv://`)
3. **Saved:** Did you click Save after adding?
4. **Redeployed:** Did you redeploy after adding?

### Alternative: Use Render Shell
1. Go to your service → **Shell** tab
2. Run: `echo $MONGODB_URI`
3. If it's empty, the variable isn't set correctly

## MongoDB Atlas Network Access (CRITICAL!)

Even with the correct environment variable, you MUST allow Render to connect:

1. Go to: https://cloud.mongodb.com
2. Click **Network Access** (left sidebar)
3. Click **Add IP Address**
4. Click **Allow Access from Anywhere** button
5. Click **Confirm**

This adds `0.0.0.0/0` which allows all IPs (including Render's).

## Still Not Working?

If after following all steps you still see the warning:

1. **Screenshot your Environment Variables page** and check:
   - Is `MONGODB_URI` listed?
   - What does the value show?

2. **Check Render service type:**
   - Is it a "Web Service"?
   - Is it connected to your GitHub repo?

3. **Try deleting and re-adding:**
   - Delete the `MONGODB_URI` variable
   - Add it again with the exact value

4. **Contact Render support** with:
   - Your service name
   - Screenshot of Environment Variables
   - The error logs

