#!/bin/bash
# Run database migrations on Render

set -e

if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL environment variable is not set"
    exit 1
fi

echo "Running migrations..."

echo "Running migration 001_initial_schema.up.sql..."
psql "$DATABASE_URL" < migrations/001_initial_schema.up.sql

echo "Running migration 002_mood_tracking.up.sql..."
psql "$DATABASE_URL" < migrations/002_mood_tracking.up.sql

echo "Running migration 003_add_quiz_questions.up.sql..."
psql "$DATABASE_URL" < migrations/003_add_quiz_questions.up.sql

echo "✅ All migrations completed successfully!"

