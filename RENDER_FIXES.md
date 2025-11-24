# Render Deployment Fixes

## Issues Fixed

### 1. Session Cookies (401 Errors)
**Problem**: Cookies weren't being set properly in production (HTTPS)

**Fix**: Updated session cookie configuration:
- Set `SameSite: Lax` mode (works with HTTPS)
- Auto-detect HTTPS and set `Secure: true`
- Proper cookie path and max age

### 2. Template Loading (500 Errors)
**Problem**: Templates might not load correctly in production

**Fix**: 
- Use template engine instead of direct file serve
- Added fallback to file serve if template fails
- Better error handling

### 3. Static Files
**Problem**: Static files need to be accessible

**Fix**: Ensure `static/` directory is committed and served correctly

## Deployment Checklist

1. ✅ Code pushed to GitHub
2. ✅ Database created on Render
3. ✅ Web service created
4. ✅ Environment variables set:
   - `DATABASE_URL` (Internal URL)
   - `SESSION_SECRET` (generated)
   - `LLM_API_KEY` (optional)
   - `ENABLE_TRACING=false`
5. ✅ Migrations run
6. ✅ **Pull latest code** (includes cookie fixes)
7. ✅ **Redeploy** on Render

## After Redeploy

1. Clear browser cache
2. Test signup/login
3. Verify cookies are being set (check browser DevTools → Application → Cookies)
4. Test all features

## Debugging Tips

**Check Render Logs:**
- Go to your service → Logs
- Look for errors during startup
- Check for template loading errors

**Check Browser Console:**
- F12 → Console tab
- Look for JavaScript errors
- Check Network tab for failed requests

**Verify Cookies:**
- F12 → Application → Cookies
- Should see `feedback-session` cookie
- Check SameSite attribute (should be Lax)

