# Gemini AI Setup Guide

## Configuration

To use Google Gemini AI instead of OpenAI, set these environment variables in Render:

### Option 1: Using GEMINI_* variables (Recommended)
- **GEMINI_API_URL**: `https://generativelanguage.googleapis.com/v1beta`
- **GEMINI_API_KEY**: Your Gemini API key
- **GEMINI_MODEL**: `gemini-pro` (or `gemini-1.5-pro`)

### Option 2: Using LLM_* variables
- **LLM_API_URL**: `https://generativelanguage.googleapis.com/v1beta`
- **LLM_API_KEY**: Your Gemini API key  
- **LLM_MODEL**: `gemini-pro`

## Getting a Gemini API Key

1. Go to https://makersuite.google.com/app/apikey
2. Sign in with your Google account
3. Click "Create API Key"
4. Copy the API key
5. Set it in Render dashboard as `GEMINI_API_KEY` or `LLM_API_KEY`

## Features Added

✅ **Gemini API Support** - Full compatibility with Google Gemini
✅ **Loading Indicator** - Shows "Counselor is typing..." while waiting for response
✅ **Enhanced Daily Mood Assessment** - Now has 8 questions instead of 1:
   - Overall mood
   - Energy level
   - Sleep quality
   - Concentration ability
   - Social connection
   - Stress level
   - Optimism
   - Emotional well-being

## Troubleshooting

**No response from Gemini:**
- Check API key is correct
- Verify API URL is: `https://generativelanguage.googleapis.com/v1beta`
- Check model name (use `gemini-pro` or `gemini-1.5-pro`)
- Check Render logs for error messages

**Loading indicator not showing:**
- Clear browser cache
- Check browser console for JavaScript errors

