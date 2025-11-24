# Render Deployment Setup with MongoDB

## Step 1: Set Environment Variable in Render

1. Go to your Render dashboard: https://dashboard.render.com
2. Select your `feedback-sys` service
3. Go to **Environment** tab
4. Click **Add Environment Variable**
5. Add:
   - **Key:** `MONGODB_URI`
   - **Value:** `mongodb+srv://joan_test:Redwater710@cluster0.kniyd8u.mongodb.net/feedback_sys?retryWrites=true&w=majority`
6. Click **Save Changes**

## Step 2: Verify MongoDB Atlas Network Access

1. Go to MongoDB Atlas: https://cloud.mongodb.com
2. Select your cluster
3. Click **Network Access** (left sidebar)
4. Click **Add IP Address**
5. Click **Allow Access from Anywhere** (or add Render's IP ranges)
6. Click **Confirm**

## Step 3: Deploy

1. Push your code to GitHub:
   ```bash
   git add .
   git commit -m "Configure MongoDB Atlas connection"
   git push origin main
   ```

2. Render will automatically deploy

3. Check logs in Render dashboard to verify connection

## Step 4: Test

1. Visit your Render URL (e.g., `https://feedback-sys-2zmw.onrender.com`)
2. Try signing up - should work without 500 errors!
3. Try logging in - should work without 401 errors!

## Troubleshooting

**Connection Error:**
- Check MongoDB Atlas Network Access allows all IPs (0.0.0.0/0)
- Verify password is correct
- Check connection string format

**500 Error on Signup:**
- Check Render logs for detailed error
- Verify MONGODB_URI is set correctly
- Ensure database name is `feedback_sys`

