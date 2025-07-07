-- Test Children Seed Data for MamaCare SL
-- Sample children data for development and testing only, not for production

-- Clear existing data if needed (commented out to avoid accidental deletion)
-- TRUNCATE TABLE children RESTART IDENTITY CASCADE;

-- Insert test children
INSERT INTO children
(first_name, middle_name, last_name, date_of_birth, birth_weight_grams, birth_length_cm, 
 place_of_birth, delivery_type, gestational_age_weeks, mother_id, father_name, 
 blood_type, current_growth_status)
VALUES
-- Aminata's children
('Ibrahim', NULL, 'Kamara', '2021-05-15', 3200, 50, 
 'Princess Christian Maternity Hospital', 'VAGINAL', 39, 
 (SELECT id FROM users WHERE phone_number = '+23277000001'), 'Ibrahim Kamara Sr.', 
 'O+', 'NORMAL'),
 
('Hawa', NULL, 'Kamara', '2019-07-20', 2900, 48, 
 'Home', 'VAGINAL', 37, 
 (SELECT id FROM users WHERE phone_number = '+23277000001'), 'Ibrahim Kamara Sr.', 
 'O+', 'UNDERWEIGHT'),

-- Fatmata's child
('Abdul', 'Rahman', 'Sesay', '2022-01-10', 3500, 52, 
 'Princess Christian Maternity Hospital', 'C_SECTION', 38, 
 (SELECT id FROM users WHERE phone_number = '+23277000002'), 'Mohammed Sesay', 
 'A+', 'NORMAL'),

-- Isatu's children
('Amie', NULL, 'Turay', '2022-03-25', 3100, 49, 
 'Bo Government Hospital', 'VAGINAL', 40, 
 (SELECT id FROM users WHERE phone_number = '+23277000004'), 'Sorie Turay', 
 'AB+', 'NORMAL'),
 
('Mohamed', NULL, 'Turay', '2020-11-12', 2800, 47, 
 'Home', 'VAGINAL', 36, 
 (SELECT id FROM users WHERE phone_number = '+23277000004'), 'Sorie Turay', 
 'O+', 'STUNTED'),

-- Kadiatu's child
('Foday', NULL, 'Bangura', '2022-06-01', 3300, 51, 
 'Connaught Hospital', 'VAGINAL', 41, 
 (SELECT id FROM users WHERE phone_number = '+23277000005'), 'Amadu Bangura', 
 'O-', 'NORMAL');

-- Insert growth measurements for Ibrahim Kamara
INSERT INTO growth_measurements
(child_id, measured_at, weight_grams, height_cm, head_circumference_cm, 
 weight_for_age_z, height_for_age_z, weight_for_height_z, growth_status,
 measured_by_user_id, measured_at_facility_id)
VALUES
((SELECT id FROM children WHERE first_name = 'Ibrahim' AND last_name = 'Kamara'), 
 '2021-06-15', 4200, 54, 38.5, 0.2, 0.1, 0.3, 'NORMAL',
 (SELECT id FROM users WHERE role = 'CHW' LIMIT 1),
 (SELECT id FROM healthcare_facilities WHERE name = 'Connaught Hospital' LIMIT 1)),
 
((SELECT id FROM children WHERE first_name = 'Ibrahim' AND last_name = 'Kamara'), 
 '2021-09-15', 6500, 62, 41.2, 0.3, 0.2, 0.4, 'NORMAL',
 (SELECT id FROM users WHERE role = 'CHW' LIMIT 1),
 (SELECT id FROM healthcare_facilities WHERE name = 'Connaught Hospital' LIMIT 1)),
 
((SELECT id FROM children WHERE first_name = 'Ibrahim' AND last_name = 'Kamara'), 
 '2022-05-15', 9200, 74, 45.8, 0.1, 0.0, 0.2, 'NORMAL',
 (SELECT id FROM users WHERE role = 'CHW' LIMIT 1),
 (SELECT id FROM healthcare_facilities WHERE name = 'Connaught Hospital' LIMIT 1));

-- Insert growth measurements for Hawa Kamara (underweight)
INSERT INTO growth_measurements
(child_id, measured_at, weight_grams, height_cm, head_circumference_cm, 
 weight_for_age_z, height_for_age_z, weight_for_height_z, growth_status,
 measured_by_user_id, measured_at_facility_id)
VALUES
((SELECT id FROM children WHERE first_name = 'Hawa' AND last_name = 'Kamara'), 
 '2022-07-20', 8900, 82, 47.0, -1.8, -0.5, -2.1, 'UNDERWEIGHT',
 (SELECT id FROM users WHERE role = 'CHW' LIMIT 1),
 (SELECT id FROM healthcare_facilities WHERE name = 'Connaught Hospital' LIMIT 1)),
 
((SELECT id FROM children WHERE first_name = 'Hawa' AND last_name = 'Kamara'), 
 '2022-10-20', 9300, 85, 47.5, -1.9, -0.6, -2.2, 'UNDERWEIGHT',
 (SELECT id FROM users WHERE role = 'CHW' LIMIT 1),
 (SELECT id FROM healthcare_facilities WHERE name = 'Connaught Hospital' LIMIT 1));

-- Insert immunization records for Ibrahim Kamara
INSERT INTO immunization_records
(child_id, vaccine_name, vaccine_dose, administered_date, 
 administered_by_user_id, administered_at_facility_id, batch_number)
VALUES
((SELECT id FROM children WHERE first_name = 'Ibrahim' AND last_name = 'Kamara'),
 'BCG', '1', '2021-05-15', 
 (SELECT id FROM users WHERE role = 'CLINICIAN' LIMIT 1),
 (SELECT id FROM healthcare_facilities WHERE name = 'Princess Christian Maternity Hospital' LIMIT 1),
 'BCG-2021-05-A'),
 
((SELECT id FROM children WHERE first_name = 'Ibrahim' AND last_name = 'Kamara'),
 'OPV', '0', '2021-05-15', 
 (SELECT id FROM users WHERE role = 'CLINICIAN' LIMIT 1),
 (SELECT id FROM healthcare_facilities WHERE name = 'Princess Christian Maternity Hospital' LIMIT 1),
 'OPV-2021-05-A'),
 
((SELECT id FROM children WHERE first_name = 'Ibrahim' AND last_name = 'Kamara'),
 'Hepatitis B', '1', '2021-05-15', 
 (SELECT id FROM users WHERE role = 'CLINICIAN' LIMIT 1),
 (SELECT id FROM healthcare_facilities WHERE name = 'Princess Christian Maternity Hospital' LIMIT 1),
 'HEP-2021-05-A'),
 
((SELECT id FROM children WHERE first_name = 'Ibrahim' AND last_name = 'Kamara'),
 'OPV', '1', '2021-07-01', 
 (SELECT id FROM users WHERE role = 'CLINICIAN' LIMIT 1),
 (SELECT id FROM healthcare_facilities WHERE name = 'Connaught Hospital' LIMIT 1),
 'OPV-2021-07-A'),
 
((SELECT id FROM children WHERE first_name = 'Ibrahim' AND last_name = 'Kamara'),
 'Penta', '1', '2021-07-01', 
 (SELECT id FROM users WHERE role = 'CLINICIAN' LIMIT 1),
 (SELECT id FROM healthcare_facilities WHERE name = 'Connaught Hospital' LIMIT 1),
 'PENTA-2021-07-A');

-- Create sample visits for prenatal care (Aminata Kamara)
INSERT INTO visits
(mother_id, visit_type, status, scheduled_date, facility_id, 
 reminder_sent, completed_date, completed_by_user_id, notes)
VALUES
((SELECT id FROM users WHERE phone_number = '+23277000001'), 
 'ANTENATAL', 'COMPLETED', '2023-03-08', 
 (SELECT id FROM healthcare_facilities WHERE name = 'Princess Christian Maternity Hospital' LIMIT 1),
 TRUE, '2023-03-08', 
 (SELECT id FROM users WHERE role = 'CLINICIAN' LIMIT 1),
 'First antenatal visit. Normal findings. Iron supplements prescribed.'),
 
((SELECT id FROM users WHERE phone_number = '+23277000001'), 
 'ANTENATAL', 'COMPLETED', '2023-04-12', 
 (SELECT id FROM healthcare_facilities WHERE name = 'Princess Christian Maternity Hospital' LIMIT 1),
 TRUE, '2023-04-13', 
 (SELECT id FROM users WHERE role = 'CLINICIAN' LIMIT 1),
 'Follow-up visit. Blood pressure slightly elevated. Advised rest and diet modifications.'),
 
((SELECT id FROM users WHERE phone_number = '+23277000001'), 
 'ANTENATAL', 'COMPLETED', '2023-05-17', 
 (SELECT id FROM healthcare_facilities WHERE name = 'Princess Christian Maternity Hospital' LIMIT 1),
 TRUE, '2023-05-17', 
 (SELECT id FROM users WHERE role = 'CLINICIAN' LIMIT 1),
 'Routine checkup. Normal findings. Fetal development on track.'),
 
((SELECT id FROM users WHERE phone_number = '+23277000001'), 
 'ANTENATAL', 'SCHEDULED', '2023-06-21', 
 (SELECT id FROM healthcare_facilities WHERE name = 'Princess Christian Maternity Hospital' LIMIT 1),
 FALSE, NULL, NULL, NULL);

-- Create vaccination visits for children
INSERT INTO visits
(child_id, visit_type, status, scheduled_date, facility_id, 
 reminder_sent, completed_date, completed_by_user_id, notes)
VALUES
((SELECT id FROM children WHERE first_name = 'Abdul' AND last_name = 'Sesay'), 
 'VACCINATION', 'COMPLETED', '2022-03-10', 
 (SELECT id FROM healthcare_facilities WHERE name = 'Connaught Hospital' LIMIT 1),
 TRUE, '2022-03-10', 
 (SELECT id FROM users WHERE role = 'CLINICIAN' LIMIT 1),
 'OPV and Penta vaccines administered. No adverse reactions.'),
 
((SELECT id FROM children WHERE first_name = 'Abdul' AND last_name = 'Sesay'), 
 'VACCINATION', 'SCHEDULED', '2022-05-12', 
 (SELECT id FROM healthcare_facilities WHERE name = 'Connaught Hospital' LIMIT 1),
 FALSE, NULL, NULL, NULL),
 
((SELECT id FROM children WHERE first_name = 'Amie' AND last_name = 'Turay'), 
 'VACCINATION', 'MISSED', '2022-05-15', 
 (SELECT id FROM healthcare_facilities WHERE name = 'Bo Government Hospital' LIMIT 1),
 TRUE, NULL, NULL, 'Mother called to reschedule due to transportation issues.');

-- Create sample SOS event
INSERT INTO sos_events
(user_id, emergency_type, description, status, for_self, location, 
 location_address, location_description, responding_facility_id, 
 responding_user_id, ambulance_dispatched, notes)
VALUES
((SELECT id FROM users WHERE phone_number = '+23277000004'), 
 'MEDICAL', 'Severe headache and blurred vision', 'CLOSED', TRUE, 
 ST_GeomFromText('POINT(-11.7380 7.9590)', 4326)::geography,
 'Near Bo Government Hospital', 'White house with blue roof', 
 (SELECT id FROM healthcare_facilities WHERE name = 'Bo Government Hospital' LIMIT 1),
 (SELECT id FROM users WHERE role = 'CLINICIAN' LIMIT 1), 
 TRUE, 'Patient arrived at hospital and received treatment for pre-eclampsia.');

-- Insert sample ambulance
INSERT INTO ambulances
(vehicle_id, name, status, facility_id, vehicle_type, capacity, 
 has_oxygen, has_stretcher, current_driver_name, current_driver_phone)
VALUES
('SL-AMB-001', 'Ambulance 1', 'AVAILABLE', 
 (SELECT id FROM healthcare_facilities WHERE name = 'Connaught Hospital' LIMIT 1),
 'AMBULANCE', 3, TRUE, TRUE, 'James Wilson', '+23278901234'),
 
('SL-AMB-002', 'Ambulance 2', 'AVAILABLE', 
 (SELECT id FROM healthcare_facilities WHERE name = 'Princess Christian Maternity Hospital' LIMIT 1),
 'AMBULANCE', 3, TRUE, TRUE, 'Samuel Johnson', '+23278901235');

-- Sample screener results
INSERT INTO screener_results
(mother_id, screened_at, screened_by_user_id, completed, 
 risk_level, primary_concern, danger_signs_detected, 
 followup_recommended, action_taken)
VALUES
((SELECT id FROM users WHERE phone_number = '+23277000001'),
 '2023-05-01 10:15:00', 
 (SELECT id FROM users WHERE role = 'CHW' LIMIT 1),
 TRUE, 'YELLOW', 'Headache and swelling', FALSE, 
 TRUE, 'Advised to visit PCMH for checkup within 48 hours.');

-- Insert corresponding screener answers
INSERT INTO screener_answers
(screener_result_id, question_id, question_text, answer_boolean, answer_sequence, contributed_to_risk, individual_risk_level)
VALUES
((SELECT id FROM screener_results ORDER BY screened_at DESC LIMIT 1),
 (SELECT id FROM screener_questions WHERE question_code = 'MATERNAL_BLEEDING'),
 'Are you experiencing any vaginal bleeding?', 
 FALSE, 1, FALSE, NULL),
 
((SELECT id FROM screener_results ORDER BY screened_at DESC LIMIT 1),
 (SELECT id FROM screener_questions WHERE question_code = 'MATERNAL_HEADACHE'),
 'Do you have a persistent headache?', 
 TRUE, 2, TRUE, 'YELLOW'),
 
((SELECT id FROM screener_results ORDER BY screened_at DESC LIMIT 1),
 (SELECT id FROM screener_questions WHERE question_code = 'MATERNAL_SWELLING'),
 'Do you have swelling of the face, hands, or feet?', 
 TRUE, 3, TRUE, 'YELLOW');
