-- Mood tracking tables
CREATE TABLE mood_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mood_type VARCHAR(50) NOT NULL,
    mood_level VARCHAR(50) NOT NULL,
    score INTEGER NOT NULL CHECK (score >= 1 AND score <= 10),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    date DATE NOT NULL DEFAULT CURRENT_DATE
);

CREATE INDEX idx_mood_entries_user_id ON mood_entries(user_id);
CREATE INDEX idx_mood_entries_date ON mood_entries(date);
CREATE INDEX idx_mood_entries_user_date ON mood_entries(user_id, date);
CREATE UNIQUE INDEX idx_mood_entries_user_date_unique ON mood_entries(user_id, date);

-- Mood recommendations
CREATE TABLE mood_recommendations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mood_entry_id UUID NOT NULL REFERENCES mood_entries(id) ON DELETE CASCADE,
    recommendations TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_mood_recommendations_user_id ON mood_recommendations(user_id);
CREATE INDEX idx_mood_recommendations_mood_entry_id ON mood_recommendations(mood_entry_id);

-- Quiz tables
CREATE TABLE quizzes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_quizzes_type ON quizzes(type);

-- Quiz questions
CREATE TABLE quiz_questions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    question TEXT NOT NULL,
    options JSONB, -- Array of options for multiple choice
    question_type VARCHAR(50) NOT NULL DEFAULT 'multiple_choice',
    "order" INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_quiz_questions_quiz_id ON quiz_questions(quiz_id);
CREATE INDEX idx_quiz_questions_order ON quiz_questions(quiz_id, "order");

-- Quiz responses
CREATE TABLE quiz_responses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    answers JSONB NOT NULL, -- Map of question_id -> answer
    score INTEGER NOT NULL,
    result TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_quiz_responses_user_id ON quiz_responses(user_id);
CREATE INDEX idx_quiz_responses_quiz_id ON quiz_responses(quiz_id);
CREATE INDEX idx_quiz_responses_created_at ON quiz_responses(created_at);

-- Quiz recommendations
CREATE TABLE quiz_recommendations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quiz_response_id UUID NOT NULL REFERENCES quiz_responses(id) ON DELETE CASCADE,
    recommendations TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_quiz_recommendations_user_id ON quiz_recommendations(user_id);
CREATE INDEX idx_quiz_recommendations_quiz_response_id ON quiz_recommendations(quiz_response_id);

-- Motivational quotes
CREATE TABLE motivational_quotes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quote TEXT NOT NULL,
    author VARCHAR(255),
    mood_type VARCHAR(50),
    mood_level VARCHAR(50),
    is_ai BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    date DATE NOT NULL DEFAULT CURRENT_DATE
);

CREATE INDEX idx_motivational_quotes_user_id ON motivational_quotes(user_id);
CREATE INDEX idx_motivational_quotes_date ON motivational_quotes(date);
CREATE INDEX idx_motivational_quotes_user_date ON motivational_quotes(user_id, date);

-- Insert default quizzes
INSERT INTO quizzes (id, type, title, description) VALUES
    (uuid_generate_v4(), 'mood_assessment', 'Daily Mood Assessment', 'A quick assessment to understand your current mood and emotional state'),
    (uuid_generate_v4(), 'stress_level', 'Stress Level Check', 'Evaluate your current stress levels and identify potential stressors'),
    (uuid_generate_v4(), 'anxiety_check', 'Anxiety Check', 'Assess your anxiety levels and get personalized recommendations'),
    (uuid_generate_v4(), 'wellness', 'Wellness Check', 'Comprehensive wellness assessment covering multiple aspects of your mental health');

-- Insert default questions for mood assessment quiz
DO $$
DECLARE
    mood_quiz_id UUID;
BEGIN
    SELECT id INTO mood_quiz_id FROM quizzes WHERE type = 'mood_assessment' LIMIT 1;
    
    INSERT INTO quiz_questions (quiz_id, question, options, question_type, "order") VALUES
        (mood_quiz_id, 'How would you rate your overall mood today?', 
         '["Very Low (1-2)", "Low (3-4)", "Moderate (5-6)", "Good (7-8)", "Excellent (9-10)"]'::jsonb, 
         'multiple_choice', 1),
        (mood_quiz_id, 'What best describes your current emotional state?', 
         '["Happy", "Sad", "Anxious", "Stressed", "Calm", "Energetic", "Tired", "Frustrated", "Neutral"]'::jsonb, 
         'multiple_choice', 2),
        (mood_quiz_id, 'How well did you sleep last night?', 
         '["Very Poor", "Poor", "Fair", "Good", "Excellent"]'::jsonb, 
         'multiple_choice', 3),
        (mood_quiz_id, 'How would you rate your energy levels today?', 
         '["Very Low", "Low", "Moderate", "High", "Very High"]'::jsonb, 
         'multiple_choice', 4),
        (mood_quiz_id, 'How would you rate your ability to focus today?', 
         '["Very Poor", "Poor", "Fair", "Good", "Excellent"]'::jsonb, 
         'multiple_choice', 5);
END $$;

