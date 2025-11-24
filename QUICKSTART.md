# Quick Start Guide

## Prerequisites

- Go 1.21+
- PostgreSQL 12+
- LLM API key (OpenAI-compatible)

## Quick Setup

1. **Start PostgreSQL and Jaeger (optional)**:
   ```bash
   docker-compose up -d
   ```

2. **Set up database**:
   ```bash
   createdb feedback_sys
   psql feedback_sys < migrations/001_initial_schema.up.sql
   ```

3. **Configure environment**:
   ```bash
   cp .env.example .env
   # Edit .env and set:
   # - DATABASE_URL
   # - SESSION_SECRET (use a strong random string)
   # - LLM_API_KEY (your OpenAI API key)
   ```

4. **Install dependencies and run**:
   ```bash
   go mod download
   go run cmd/server/main.go
   ```

5. **Access the application**:
   Open http://localhost:8080 in your browser

## Testing the Application

1. **Sign Up**: Visit `/signup` and create an anonymous account
2. **Save Your Token**: After signup, save the token shown (or check browser console)
3. **Submit Feedback**: Go to dashboard and submit anonymous feedback
4. **Consultation**: Visit `/consultation` to chat with the AI counselor

## Anonymous Token System

- Users receive a unique cryptographic token upon signup
- This token is stored in the session cookie
- Users can save their token to access their account from other devices
- No email or phone number required!

## Troubleshooting

- **Database connection error**: Ensure PostgreSQL is running and DATABASE_URL is correct
- **LLM API error**: Verify LLM_API_KEY is set correctly
- **Port already in use**: Change PORT in .env file

