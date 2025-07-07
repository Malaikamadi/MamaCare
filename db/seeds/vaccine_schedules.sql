-- Vaccine Schedules Seed Data for MamaCare SL
-- Based on WHO Expanded Programme on Immunization (EPI) for Sierra Leone

-- Clear existing data (for re-seeding)
TRUNCATE TABLE vaccine_schedules RESTART IDENTITY CASCADE;

-- Insert vaccine schedule based on Sierra Leone's EPI
INSERT INTO vaccine_schedules 
(vaccine_name, vaccine_dose, age_months, window_start_days, window_end_days, disease_target, is_required)
VALUES
-- At Birth
('BCG', '1', 0, 0, 30, 'Tuberculosis', TRUE),
('OPV', '0', 0, 0, 30, 'Polio', TRUE),
('Hepatitis B', '1', 0, 0, 30, 'Hepatitis B', TRUE),

-- 6 Weeks (1.5 months)
('OPV', '1', 1.5, -7, 30, 'Polio', TRUE),
('Penta', '1', 1.5, -7, 30, 'Diphtheria, Pertussis, Tetanus, Hepatitis B, Hib', TRUE),
('PCV', '1', 1.5, -7, 30, 'Pneumococcal disease', TRUE),
('Rotavirus', '1', 1.5, -7, 30, 'Rotavirus diarrhea', TRUE),

-- 10 Weeks (2.5 months)
('OPV', '2', 2.5, -7, 30, 'Polio', TRUE),
('Penta', '2', 2.5, -7, 30, 'Diphtheria, Pertussis, Tetanus, Hepatitis B, Hib', TRUE),
('PCV', '2', 2.5, -7, 30, 'Pneumococcal disease', TRUE),
('Rotavirus', '2', 2.5, -7, 30, 'Rotavirus diarrhea', TRUE),

-- 14 Weeks (3.5 months)
('OPV', '3', 3.5, -7, 30, 'Polio', TRUE),
('IPV', '1', 3.5, -7, 30, 'Polio', TRUE),
('Penta', '3', 3.5, -7, 30, 'Diphtheria, Pertussis, Tetanus, Hepatitis B, Hib', TRUE),
('PCV', '3', 3.5, -7, 30, 'Pneumococcal disease', TRUE),

-- 9 Months
('Measles/Rubella', '1', 9, -7, 60, 'Measles and Rubella', TRUE),
('Yellow Fever', '1', 9, -7, 60, 'Yellow Fever', TRUE),
('Vitamin A', '1', 9, -7, 60, 'Vitamin A deficiency', TRUE),

-- 15 Months
('Measles/Rubella', '2', 15, -7, 60, 'Measles and Rubella', TRUE),
('Meningitis A', '1', 15, -7, 60, 'Meningitis A', TRUE),

-- 18 Months
('DTP', 'Booster', 18, -14, 60, 'Diphtheria, Tetanus, Pertussis', TRUE),
('OPV', 'Booster', 18, -14, 60, 'Polio', TRUE),
('Vitamin A', '2', 18, -14, 60, 'Vitamin A deficiency', TRUE),

-- Other boosters
('Tetanus toxoid', 'Pregnant women', NULL, NULL, NULL, 'Maternal and Neonatal Tetanus', TRUE);

-- Update the window values for "catch-up" immunizations that can be given later if missed
UPDATE vaccine_schedules
SET window_end_days = 365
WHERE vaccine_name IN ('BCG', 'Hepatitis B', 'Measles/Rubella');
