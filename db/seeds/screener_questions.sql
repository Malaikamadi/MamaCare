-- Screener Questions Seed Data for MamaCare SL
-- Contains health screening questions for risk assessment

-- Clear existing data (for re-seeding)
TRUNCATE TABLE screener_questions RESTART IDENTITY CASCADE;

-- Insert maternal danger sign questions
INSERT INTO screener_questions 
(question_text, question_code, answer_type, is_danger_sign, risk_category, category, subcategory, display_order, translation_key)
VALUES
-- Severe danger signs (RED risk level)
('Are you experiencing any vaginal bleeding?', 'MATERNAL_BLEEDING', 'BOOLEAN', TRUE, 'RED', 'Maternal', 'Bleeding', 1, 'question.maternal.bleeding'),
('Have you had any convulsions or fits?', 'MATERNAL_CONVULSIONS', 'BOOLEAN', TRUE, 'RED', 'Maternal', 'Neurological', 2, 'question.maternal.convulsions'),
('Do you have severe abdominal pain?', 'MATERNAL_SEVERE_PAIN', 'BOOLEAN', TRUE, 'RED', 'Maternal', 'Pain', 3, 'question.maternal.severe_pain'),
('Do you have fever with severe headache or neck stiffness?', 'MATERNAL_FEVER_HEADACHE', 'BOOLEAN', TRUE, 'RED', 'Maternal', 'Infection', 4, 'question.maternal.fever_headache'),
('Are you unable to keep food or fluids down for more than 24 hours?', 'MATERNAL_DEHYDRATION', 'BOOLEAN', TRUE, 'RED', 'Maternal', 'Dehydration', 5, 'question.maternal.dehydration'),

-- Moderate risk signs (YELLOW risk level)
('Do you have a persistent headache?', 'MATERNAL_HEADACHE', 'BOOLEAN', FALSE, 'YELLOW', 'Maternal', 'Neurological', 6, 'question.maternal.headache'),
('Do you have swelling of the face, hands, or feet?', 'MATERNAL_SWELLING', 'BOOLEAN', FALSE, 'YELLOW', 'Maternal', 'Preeclampsia', 7, 'question.maternal.swelling'),
('Do you have blurred vision or see spots?', 'MATERNAL_VISION', 'BOOLEAN', FALSE, 'YELLOW', 'Maternal', 'Preeclampsia', 8, 'question.maternal.vision'),
('Have you felt a reduction in your baby''s movements?', 'MATERNAL_REDUCED_MOVEMENT', 'BOOLEAN', FALSE, 'YELLOW', 'Maternal', 'Fetal', 9, 'question.maternal.reduced_movement'),
('Do you have a fever?', 'MATERNAL_FEVER', 'BOOLEAN', FALSE, 'YELLOW', 'Maternal', 'Infection', 10, 'question.maternal.fever'),
('Do you have burning when urinating?', 'MATERNAL_UTI', 'BOOLEAN', FALSE, 'YELLOW', 'Maternal', 'Infection', 11, 'question.maternal.uti'),

-- Low risk questions (information gathering)
('How many weeks pregnant are you?', 'MATERNAL_GESTATION', 'NUMERIC', FALSE, NULL, 'Maternal', 'General', 12, 'question.maternal.gestation'),
('Have you attended any antenatal care visits?', 'MATERNAL_ANC', 'BOOLEAN', FALSE, NULL, 'Maternal', 'General', 13, 'question.maternal.anc'),
('Are you taking iron and folic acid supplements?', 'MATERNAL_SUPPLEMENTS', 'BOOLEAN', FALSE, NULL, 'Maternal', 'Nutrition', 14, 'question.maternal.supplements'),
('How many meals do you eat per day?', 'MATERNAL_NUTRITION', 'NUMERIC', FALSE, NULL, 'Maternal', 'Nutrition', 15, 'question.maternal.nutrition');

-- Insert newborn/infant danger sign questions
INSERT INTO screener_questions 
(question_text, question_code, answer_type, is_danger_sign, risk_category, category, subcategory, display_order, translation_key)
VALUES
-- Severe danger signs (RED risk level)
('Is the child having difficulty breathing?', 'CHILD_BREATHING', 'BOOLEAN', TRUE, 'RED', 'Child', 'Respiratory', 16, 'question.child.breathing'),
('Does the child have convulsions or fits?', 'CHILD_CONVULSIONS', 'BOOLEAN', TRUE, 'RED', 'Child', 'Neurological', 17, 'question.child.convulsions'),
('Is the child unable to feed or drink?', 'CHILD_FEEDING', 'BOOLEAN', TRUE, 'RED', 'Child', 'Nutrition', 18, 'question.child.feeding'),
('Does the child have a high fever (above 38.5°C)?', 'CHILD_HIGH_FEVER', 'BOOLEAN', TRUE, 'RED', 'Child', 'Infection', 19, 'question.child.high_fever'),
('Is the child lethargic or unconscious?', 'CHILD_LETHARGY', 'BOOLEAN', TRUE, 'RED', 'Child', 'Neurological', 20, 'question.child.lethargy'),

-- Moderate risk signs (YELLOW risk level)
('Does the child have diarrhea?', 'CHILD_DIARRHEA', 'BOOLEAN', FALSE, 'YELLOW', 'Child', 'Gastrointestinal', 21, 'question.child.diarrhea'),
('Has the child been vomiting?', 'CHILD_VOMITING', 'BOOLEAN', FALSE, 'YELLOW', 'Child', 'Gastrointestinal', 22, 'question.child.vomiting'),
('Does the child have a rash?', 'CHILD_RASH', 'BOOLEAN', FALSE, 'YELLOW', 'Child', 'Skin', 23, 'question.child.rash'),
('Does the child have a cough?', 'CHILD_COUGH', 'BOOLEAN', FALSE, 'YELLOW', 'Child', 'Respiratory', 24, 'question.child.cough'),
('Does the child have an ear infection or discharge?', 'CHILD_EAR', 'BOOLEAN', FALSE, 'YELLOW', 'Child', 'Infection', 25, 'question.child.ear'),

-- Low risk questions (information gathering)
('How old is the child?', 'CHILD_AGE', 'NUMERIC', FALSE, NULL, 'Child', 'General', 26, 'question.child.age'),
('Is the child up-to-date with vaccinations?', 'CHILD_VACCINATION', 'BOOLEAN', FALSE, NULL, 'Child', 'General', 27, 'question.child.vaccination'),
('Is the child breastfeeding?', 'CHILD_BREASTFEEDING', 'BOOLEAN', FALSE, NULL, 'Child', 'Nutrition', 28, 'question.child.breastfeeding'),
('What is the child''s weight?', 'CHILD_WEIGHT', 'NUMERIC', FALSE, NULL, 'Child', 'Growth', 29, 'question.child.weight'),
('Has the child lost weight recently?', 'CHILD_WEIGHT_LOSS', 'BOOLEAN', FALSE, NULL, 'Child', 'Growth', 30, 'question.child.weight_loss');

-- Insert multiple-choice questions
UPDATE screener_questions
SET answer_options = '["Less than 3", "3-4", "5 or more"]'
WHERE question_code = 'MATERNAL_NUTRITION';

-- Define dependencies between questions
UPDATE screener_questions
SET depends_on_question_id = (SELECT id FROM screener_questions WHERE question_code = 'CHILD_DIARRHEA'),
    depends_on_answer = 'true'
WHERE question_code = 'CHILD_WEIGHT_LOSS';

UPDATE screener_questions
SET depends_on_question_id = (SELECT id FROM screener_questions WHERE question_code = 'MATERNAL_FEVER'),
    depends_on_answer = 'true'
WHERE question_code = 'MATERNAL_FEVER_HEADACHE';
