-- Prenatal Visit Schedules Seed Data for MamaCare SL
-- Based on WHO recommendations for antenatal care

-- Clear existing data (for re-seeding)
TRUNCATE TABLE prenatal_visit_schedules RESTART IDENTITY CASCADE;

-- Insert standard prenatal visit schedule based on WHO recommendations
INSERT INTO prenatal_visit_schedules 
(visit_name, visit_number, weeks_gestation, window_start_days, window_end_days, description, is_required, danger_signs_to_check)
VALUES
-- First trimester
('First Antenatal Visit (Booking)', 1, 12, -14, 14, 
 'Initial comprehensive assessment, history taking, blood tests, HIV testing, urinalysis, blood pressure check, weight, tetanus vaccination, iron/folic acid supplements.', 
 TRUE, 
 'Vaginal bleeding, severe headache, blurred vision, fever, severe abdominal pain, difficulty breathing'),

-- Second trimester
('16-Week Checkup', 2, 16, -7, 7, 
 'Fetal growth assessment, blood pressure check, weight, urinalysis, iron/folic acid supplement review.', 
 TRUE, 
 'Vaginal bleeding, severe headache, blurred vision, fever, severe abdominal pain, reduced fetal movement'),

('20-Week Ultrasound', 3, 20, -7, 7, 
 'Detailed ultrasound scan to check fetal anatomy and development, placenta position, amniotic fluid, cervical length.', 
 TRUE, 
 'Vaginal bleeding, severe headache, blurred vision, fever, severe abdominal pain, reduced fetal movement'),

('24-Week Checkup', 4, 24, -7, 7, 
 'Fetal growth assessment, blood pressure check, weight, urinalysis, glucose screening test.', 
 TRUE, 
 'Vaginal bleeding, severe headache, blurred vision, swelling of face/hands, reduced fetal movement, contractions'),

('28-Week Checkup', 5, 28, -7, 7, 
 'Fetal growth assessment, blood pressure check, weight, urinalysis, hemoglobin test, Rh antibody test for Rh-negative mothers.', 
 TRUE, 
 'Vaginal bleeding, severe headache, blurred vision, swelling of face/hands, reduced fetal movement, contractions'),

-- Third trimester
('32-Week Checkup', 6, 32, -7, 7, 
 'Fetal growth and position assessment, blood pressure check, weight, urinalysis, review birth plan.', 
 TRUE, 
 'Vaginal bleeding, severe headache, blurred vision, swelling of face/hands, reduced fetal movement, contractions, leaking fluid'),

('36-Week Checkup', 7, 36, -7, 7, 
 'Fetal growth and position assessment, blood pressure check, weight, urinalysis, Group B streptococcus screening, discuss labor signs.', 
 TRUE, 
 'Vaginal bleeding, severe headache, blurred vision, swelling of face/hands, reduced fetal movement, contractions, leaking fluid'),

('38-Week Checkup', 8, 38, -3, 3, 
 'Fetal growth and position assessment, blood pressure check, weight, urinalysis, cervical examination if indicated.', 
 TRUE, 
 'Vaginal bleeding, severe headache, blurred vision, swelling of face/hands, reduced fetal movement, contractions, leaking fluid'),

('40-Week Checkup', 9, 40, -3, 3, 
 'Fetal wellbeing assessment, blood pressure check, weight, urinalysis, discuss induction options if pregnancy continues.', 
 TRUE, 
 'Vaginal bleeding, severe headache, blurred vision, swelling of face/hands, reduced fetal movement, contractions, leaking fluid'),

-- Post-term follow-up
('41-Week Checkup', 10, 41, -2, 2, 
 'Fetal wellbeing assessment, blood pressure check, weight, urinalysis, non-stress test, discuss induction.', 
 TRUE, 
 'Vaginal bleeding, severe headache, blurred vision, swelling of face/hands, reduced fetal movement, contractions, leaking fluid');
