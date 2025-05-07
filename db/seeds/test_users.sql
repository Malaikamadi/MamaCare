-- Test Users Seed Data for MamaCare SL
-- Sample users for development and testing only, not for production

-- Clear existing data (for re-seeding)
-- Note: This cascades to related tables, so run this with caution
TRUNCATE TABLE users RESTART IDENTITY CASCADE;

-- Insert test mothers
INSERT INTO users
(firebase_uid, phone_number, email, full_name, role, language, expected_delivery_date, last_menstrual_period, blood_type, emergency_contact_name, emergency_contact_phone)
VALUES
('test_mother_1', '+23277000001', 'mother1@test.mamacare.sl', 'Aminata Kamara', 'MOTHER', 'ENGLISH', '2023-11-15', '2023-02-08', 'O+', 'Ibrahim Kamara', '+23277000101'),
('test_mother_2', '+23277000002', 'mother2@test.mamacare.sl', 'Fatmata Sesay', 'MOTHER', 'KRIO', '2023-12-20', '2023-03-15', 'A+', 'Mohammed Sesay', '+23277000102'),
('test_mother_3', '+23277000003', 'mother3@test.mamacare.sl', 'Mariama Conteh', 'MOTHER', 'ENGLISH', NULL, NULL, 'B+', 'Abdul Conteh', '+23277000103'),
('test_mother_4', '+23277000004', 'mother4@test.mamacare.sl', 'Isatu Turay', 'MOTHER', 'MENDE', '2023-10-05', '2023-01-02', 'AB+', 'Sorie Turay', '+23277000104'),
('test_mother_5', '+23277000005', 'mother5@test.mamacare.sl', 'Kadiatu Bangura', 'MOTHER', 'TEMNE', NULL, NULL, 'O-', 'Amadu Bangura', '+23277000105');

-- Insert test CHWs (Community Health Workers)
INSERT INTO users
(firebase_uid, phone_number, email, full_name, role, language, facility_id, assigned_area)
VALUES
('test_chw_1', '+23277100001', 'chw1@test.mamacare.sl', 'Foday Koroma', 'CHW', 'ENGLISH', (SELECT id FROM healthcare_facilities WHERE name = 'Connaught Hospital' LIMIT 1), 'Freetown Central'),
('test_chw_2', '+23277100002', 'chw2@test.mamacare.sl', 'Adama Kargbo', 'CHW', 'KRIO', (SELECT id FROM healthcare_facilities WHERE name = 'Aberdeen Women''s Clinic' LIMIT 1), 'Aberdeen'),
('test_chw_3', '+23277100003', 'chw3@test.mamacare.sl', 'Samuel Koroma', 'CHW', 'ENGLISH', (SELECT id FROM healthcare_facilities WHERE name = 'Kingharman Road Hospital' LIMIT 1), 'Central Freetown');

-- Insert test clinicians
INSERT INTO users
(firebase_uid, phone_number, email, full_name, role, language, facility_id)
VALUES
('test_clinician_1', '+23277200001', 'doctor1@test.mamacare.sl', 'Dr. Sylvia Blyden', 'CLINICIAN', 'ENGLISH', (SELECT id FROM healthcare_facilities WHERE name = 'Princess Christian Maternity Hospital' LIMIT 1)),
('test_clinician_2', '+23277200002', 'doctor2@test.mamacare.sl', 'Dr. Michael Kanu', 'CLINICIAN', 'ENGLISH', (SELECT id FROM healthcare_facilities WHERE name = 'Ola During Children''s Hospital' LIMIT 1)),
('test_clinician_3', '+23277200003', 'nurse1@test.mamacare.sl', 'Nurse Sarah Johnson', 'CLINICIAN', 'ENGLISH', (SELECT id FROM healthcare_facilities WHERE name = 'Connaught Hospital' LIMIT 1));

-- Insert admin user
INSERT INTO users
(firebase_uid, phone_number, email, full_name, role, language)
VALUES
('test_admin_1', '+23277300001', 'admin@test.mamacare.sl', 'Admin User', 'ADMIN', 'ENGLISH');

-- Create notification preferences for all users
INSERT INTO notification_preferences
(user_id, enable_push_notifications, enable_sms, enable_visit_reminders, enable_vaccine_reminders, preferred_language)
SELECT id, TRUE, TRUE, TRUE, TRUE, language FROM users;

-- Set specific preferences for some users
UPDATE notification_preferences
SET enable_push_notifications = FALSE, enable_sms = TRUE
WHERE user_id = (SELECT id FROM users WHERE phone_number = '+23277000003');

UPDATE notification_preferences
SET preferred_reminder_days_before = 2
WHERE user_id = (SELECT id FROM users WHERE phone_number = '+23277000001');

-- Create device tokens for some users
INSERT INTO device_tokens
(user_id, device_id, device_name, device_platform, push_token, app_version, os_version)
VALUES
((SELECT id FROM users WHERE phone_number = '+23277000001'), 'device-uuid-1', 'Aminata''s Phone', 'ANDROID', 'ExponentPushToken[xxxxxxxxxxxxxxxxxxxx1]', '1.0.0', 'Android 11'),
((SELECT id FROM users WHERE phone_number = '+23277000002'), 'device-uuid-2', 'Fatmata''s Phone', 'ANDROID', 'ExponentPushToken[xxxxxxxxxxxxxxxxxxxx2]', '1.0.0', 'Android 10'),
((SELECT id FROM users WHERE phone_number = '+23277000004'), 'device-uuid-4', 'Isatu''s Phone', 'IOS', 'ExponentPushToken[xxxxxxxxxxxxxxxxxxxx4]', '1.0.0', 'iOS 15.0'),
((SELECT id FROM users WHERE phone_number = '+23277100001'), 'device-uuid-chw1', 'Foday''s Phone', 'ANDROID', 'ExponentPushToken[yyyyyyyyyyyyyyyyyyy1]', '1.0.0', 'Android 12'),
((SELECT id FROM users WHERE phone_number = '+23277200001'), 'device-uuid-clin1', 'Dr. Blyden''s Phone', 'IOS', 'ExponentPushToken[zzzzzzzzzzzzzzzzzzz1]', '1.0.0', 'iOS 16.0');
