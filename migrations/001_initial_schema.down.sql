-- Drop triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_feedbacks_updated_at ON feedbacks;
DROP TRIGGER IF EXISTS update_consultation_sessions_updated_at ON consultation_sessions;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS consultations;
DROP TABLE IF EXISTS consultation_sessions;
DROP TABLE IF EXISTS feedbacks;
DROP TABLE IF EXISTS users;

