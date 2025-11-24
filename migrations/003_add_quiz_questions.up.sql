-- Add questions for Stress Level Check quiz
DO $$
DECLARE
    stress_quiz_id UUID;
BEGIN
    SELECT id INTO stress_quiz_id FROM quizzes WHERE type = 'stress_level' LIMIT 1;
    
    IF stress_quiz_id IS NOT NULL THEN
        INSERT INTO quiz_questions (quiz_id, question, options, question_type, "order") VALUES
            (stress_quiz_id, 'How would you rate your overall stress level today?', 
             '["Very Low (1-2)", "Low (3-4)", "Moderate (5-6)", "High (7-8)", "Very High (9-10)"]'::jsonb, 
             'multiple_choice', 1),
            (stress_quiz_id, 'How often do you feel overwhelmed by your responsibilities?', 
             '["Never", "Rarely", "Sometimes", "Often", "Always"]'::jsonb, 
             'multiple_choice', 2),
            (stress_quiz_id, 'How well are you sleeping?', 
             '["Very Well", "Well", "Fair", "Poorly", "Very Poorly"]'::jsonb, 
             'multiple_choice', 3),
            (stress_quiz_id, 'How would you describe your ability to relax?', 
             '["Excellent", "Good", "Fair", "Poor", "Very Poor"]'::jsonb, 
             'multiple_choice', 4),
            (stress_quiz_id, 'How much control do you feel over your daily schedule?', 
             '["Complete Control", "Good Control", "Some Control", "Little Control", "No Control"]'::jsonb, 
             'multiple_choice', 5);
    END IF;
END $$;

-- Add questions for Anxiety Check quiz
DO $$
DECLARE
    anxiety_quiz_id UUID;
BEGIN
    SELECT id INTO anxiety_quiz_id FROM quizzes WHERE type = 'anxiety_check' LIMIT 1;
    
    IF anxiety_quiz_id IS NOT NULL THEN
        INSERT INTO quiz_questions (quiz_id, question, options, question_type, "order") VALUES
            (anxiety_quiz_id, 'How often do you experience feelings of anxiety or worry?', 
             '["Never", "Rarely", "Sometimes", "Often", "Constantly"]'::jsonb, 
             'multiple_choice', 1),
            (anxiety_quiz_id, 'How would you rate your anxiety level right now?', 
             '["Very Low (1-2)", "Low (3-4)", "Moderate (5-6)", "High (7-8)", "Very High (9-10)"]'::jsonb, 
             'multiple_choice', 2),
            (anxiety_quiz_id, 'Do you experience physical symptoms when anxious (racing heart, sweating, etc.)?', 
             '["Never", "Rarely", "Sometimes", "Often", "Always"]'::jsonb, 
             'multiple_choice', 3),
            (anxiety_quiz_id, 'How much does anxiety interfere with your daily activities?', 
             '["Not at All", "Slightly", "Moderately", "Significantly", "Severely"]'::jsonb, 
             'multiple_choice', 4),
            (anxiety_quiz_id, 'How would you describe your ability to manage anxious thoughts?', 
             '["Excellent", "Good", "Fair", "Poor", "Very Poor"]'::jsonb, 
             'multiple_choice', 5);
    END IF;
END $$;

-- Add questions for Wellness Check quiz
DO $$
DECLARE
    wellness_quiz_id UUID;
BEGIN
    SELECT id INTO wellness_quiz_id FROM quizzes WHERE type = 'wellness' LIMIT 1;
    
    IF wellness_quiz_id IS NOT NULL THEN
        INSERT INTO quiz_questions (quiz_id, question, options, question_type, "order") VALUES
            (wellness_quiz_id, 'How would you rate your overall mental wellness?', 
             '["Excellent (9-10)", "Good (7-8)", "Moderate (5-6)", "Poor (3-4)", "Very Poor (1-2)"]'::jsonb, 
             'multiple_choice', 1),
            (wellness_quiz_id, 'How satisfied are you with your social connections?', 
             '["Very Satisfied", "Satisfied", "Neutral", "Dissatisfied", "Very Dissatisfied"]'::jsonb, 
             'multiple_choice', 2),
            (wellness_quiz_id, 'How would you rate your physical health?', 
             '["Excellent", "Good", "Fair", "Poor", "Very Poor"]'::jsonb, 
             'multiple_choice', 3),
            (wellness_quiz_id, 'How well are you managing your academic/work responsibilities?', 
             '["Very Well", "Well", "Fairly", "Poorly", "Very Poorly"]'::jsonb, 
             'multiple_choice', 4),
            (wellness_quiz_id, 'How often do you engage in activities you enjoy?', 
             '["Daily", "Several Times a Week", "Weekly", "Rarely", "Never"]'::jsonb, 
             'multiple_choice', 5),
            (wellness_quiz_id, 'How would you rate your work-life balance?', 
             '["Excellent", "Good", "Fair", "Poor", "Very Poor"]'::jsonb, 
             'multiple_choice', 6),
            (wellness_quiz_id, 'How supported do you feel in your current environment?', 
             '["Very Supported", "Supported", "Neutral", "Unsupported", "Very Unsupported"]'::jsonb, 
             'multiple_choice', 7);
    END IF;
END $$;

