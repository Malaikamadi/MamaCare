-- Healthcare Facilities Seed Data for MamaCare SL
-- Contains healthcare facility locations across Sierra Leone

-- Clear existing data (for re-seeding)
TRUNCATE TABLE healthcare_facilities RESTART IDENTITY CASCADE;

-- Insert primary hospitals
INSERT INTO healthcare_facilities 
(name, facility_type, phone_number, email, address, district, chiefdom, location, operating_hours, has_ambulance, has_maternity_ward, bed_capacity)
VALUES
('Connaught Hospital', 'HOSPITAL', '+23222222001', 'info@connaughthospital.gov.sl', 'Lightfoot Boston Street', 'Western Area Urban', 'Freetown Central', ST_GeomFromText('POINT(-13.2343 8.4881)', 4326)::geography, '24/7', TRUE, TRUE, 300),
('Ola During Children''s Hospital', 'HOSPITAL', '+23222222002', 'info@oladuringhospital.gov.sl', 'Fourah Bay Road', 'Western Area Urban', 'East End', ST_GeomFromText('POINT(-13.2122 8.4869)', 4326)::geography, '24/7', TRUE, TRUE, 100),
('Princess Christian Maternity Hospital', 'HOSPITAL', '+23222222003', 'info@pcmh.gov.sl', 'Fourah Bay Road', 'Western Area Urban', 'East End', ST_GeomFromText('POINT(-13.2130 8.4875)', 4326)::geography, '24/7', TRUE, TRUE, 150),
('Makeni Government Hospital', 'HOSPITAL', '+23276543210', 'info@makenihospital.gov.sl', 'Hospital Road', 'Bombali', 'Makeni', ST_GeomFromText('POINT(-12.0442 8.8841)', 4326)::geography, '24/7', TRUE, TRUE, 120),
('Bo Government Hospital', 'HOSPITAL', '+23276543211', 'info@bohospital.gov.sl', 'Hospital Road, Bo', 'Bo', 'Bo Town', ST_GeomFromText('POINT(-11.7400 7.9597)', 4326)::geography, '24/7', TRUE, TRUE, 150),
('Kenema Government Hospital', 'HOSPITAL', '+23276543212', 'info@kghospital.gov.sl', 'Hangha Road', 'Kenema', 'Nongowa', ST_GeomFromText('POINT(-11.1929 7.8776)', 4326)::geography, '24/7', TRUE, TRUE, 130);

-- Insert community health centers
INSERT INTO healthcare_facilities 
(name, facility_type, phone_number, address, district, chiefdom, location, operating_hours, has_ambulance, has_maternity_ward, bed_capacity)
VALUES
('Lumley Government Hospital', 'HEALTH_CENTER', '+23276543213', 'Lumley Beach Road', 'Western Area Urban', 'Lumley', ST_GeomFromText('POINT(-13.2670 8.4700)', 4326)::geography, '8am-6pm', FALSE, TRUE, 30),
('Kingharman Road Hospital', 'HEALTH_CENTER', '+23276543214', 'Kingharman Road', 'Western Area Urban', 'Central', ST_GeomFromText('POINT(-13.2481 8.4841)', 4326)::geography, '8am-6pm', FALSE, TRUE, 25),
('Rokupa Government Hospital', 'HEALTH_CENTER', '+23276543215', 'Calaba Town Road', 'Western Area Urban', 'East End', ST_GeomFromText('POINT(-13.1910 8.4660)', 4326)::geography, '8am-6pm', FALSE, TRUE, 40),
('Waterloo Community Health Centre', 'HEALTH_CENTER', '+23276543216', 'Main Road, Waterloo', 'Western Area Rural', 'Waterloo', ST_GeomFromText('POINT(-13.0718 8.3380)', 4326)::geography, '8am-5pm', FALSE, TRUE, 20),
('Moyamba Community Health Centre', 'HEALTH_CENTER', '+23276543217', 'Hospital Road, Moyamba', 'Moyamba', 'Kaiyamba', ST_GeomFromText('POINT(-12.4318 8.1605)', 4326)::geography, '8am-5pm', FALSE, TRUE, 15),
('Kailahun Community Health Centre', 'HEALTH_CENTER', '+23276543218', 'Main Street, Kailahun', 'Kailahun', 'Luawa', ST_GeomFromText('POINT(-10.5737 8.2772)', 4326)::geography, '8am-5pm', FALSE, TRUE, 15);

-- Insert health posts
INSERT INTO healthcare_facilities 
(name, facility_type, phone_number, address, district, chiefdom, location, operating_hours, has_ambulance, has_maternity_ward, bed_capacity)
VALUES
('Congo Cross Health Post', 'HEALTH_POST', '+23276543219', 'Congo Cross', 'Western Area Urban', 'Congo Cross', ST_GeomFromText('POINT(-13.2670 8.4762)', 4326)::geography, '9am-5pm', FALSE, FALSE, 5),
('Goderich Health Post', 'HEALTH_POST', '+23276543220', 'Goderich Village', 'Western Area Rural', 'Goderich', ST_GeomFromText('POINT(-13.2929 8.4365)', 4326)::geography, '9am-5pm', FALSE, FALSE, 5),
('Hamilton Health Post', 'HEALTH_POST', '+23276543221', 'Hamilton Village', 'Western Area Rural', 'Hamilton', ST_GeomFromText('POINT(-13.2745 8.3975)', 4326)::geography, '9am-5pm', FALSE, FALSE, 5),
('Tombo Health Post', 'HEALTH_POST', '+23276543222', 'Tombo Village', 'Western Area Rural', 'Tombo', ST_GeomFromText('POINT(-13.2065 8.3066)', 4326)::geography, '9am-5pm', FALSE, FALSE, 5),
('Mile 91 Health Post', 'HEALTH_POST', '+23276543223', 'Mile 91', 'Tonkolili', 'Yoni', ST_GeomFromText('POINT(-12.2148 8.4648)', 4326)::geography, '9am-5pm', FALSE, FALSE, 5),
('Kambia Health Post', 'HEALTH_POST', '+23276543224', 'Main Street, Kambia', 'Kambia', 'Magbema', ST_GeomFromText('POINT(-12.9188 9.1261)', 4326)::geography, '9am-5pm', FALSE, FALSE, 5),
('Mattru Jong Health Post', 'HEALTH_POST', '+23276543225', 'Main Road, Mattru Jong', 'Bonthe', 'Jong', ST_GeomFromText('POINT(-12.1706 7.5848)', 4326)::geography, '9am-5pm', FALSE, FALSE, 5);

-- Insert clinics
INSERT INTO healthcare_facilities 
(name, facility_type, phone_number, email, address, district, chiefdom, location, operating_hours, has_ambulance, has_maternity_ward)
VALUES
('Aberdeen Women''s Clinic', 'CLINIC', '+23276543226', 'info@awc.org', 'Aberdeen', 'Western Area Urban', 'Aberdeen', ST_GeomFromText('POINT(-13.2878 8.4758)', 4326)::geography, '8am-6pm', FALSE, TRUE),
('Choices Community Clinic', 'CLINIC', '+23276543227', 'info@choices.org', 'Wilkinson Road', 'Western Area Urban', 'Murray Town', ST_GeomFromText('POINT(-13.2639 8.4847)', 4326)::geography, '8am-6pm', FALSE, TRUE),
('Well Woman Clinic', 'CLINIC', '+23276543228', 'info@wellwoman.org', 'Percival Street', 'Western Area Urban', 'Central', ST_GeomFromText('POINT(-13.2364 8.4841)', 4326)::geography, '8am-6pm', FALSE, TRUE),
('Bo Well Woman Clinic', 'CLINIC', '+23276543229', 'info@bowellwoman.org', 'Fenton Road, Bo', 'Bo', 'Bo Town', ST_GeomFromText('POINT(-11.7383 7.9551)', 4326)::geography, '8am-6pm', FALSE, TRUE),
('Kenema Women''s Clinic', 'CLINIC', '+23276543230', 'info@kenemawomen.org', 'Blama Road, Kenema', 'Kenema', 'Nongowa', ST_GeomFromText('POINT(-11.1892 7.8741)', 4326)::geography, '8am-6pm', FALSE, TRUE);

-- Set some facilities to have ambulances
UPDATE healthcare_facilities
SET has_ambulance = TRUE
WHERE name IN ('Connaught Hospital', 'Ola During Children''s Hospital', 'Princess Christian Maternity Hospital', 'Makeni Government Hospital', 'Bo Government Hospital', 'Kenema Government Hospital');
