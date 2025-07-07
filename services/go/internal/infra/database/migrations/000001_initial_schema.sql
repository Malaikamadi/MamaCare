-- Initial Schema Migration for MamaCare
-- Creates the core tables for the application

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS postgis;

-- Create ENUM types
CREATE TYPE user_role AS ENUM (
  'MOTHER',
  'CHW',
  'CLINICIAN',
  'ADMIN'
);

CREATE TYPE risk_level AS ENUM (
  'LOW',
  'MEDIUM',
  'HIGH'
);

CREATE TYPE facility_type AS ENUM (
  'HOSPITAL',
  'CLINIC',
  'HEALTH_CENTER',
  'MOBILE_UNIT'
);

CREATE TYPE visit_status AS ENUM (
  'SCHEDULED',
  'CHECKED_IN',
  'COMPLETED',
  'CANCELLED'
);

CREATE TYPE visit_type AS ENUM (
  'ROUTINE',
  'EMERGENCY',
  'FOLLOW_UP'
);

CREATE TYPE blood_type AS ENUM (
  'A_POSITIVE',
  'A_NEGATIVE',
  'B_POSITIVE',
  'B_NEGATIVE',
  'AB_POSITIVE',
  'AB_NEGATIVE',
  'O_POSITIVE',
  'O_NEGATIVE'
);

CREATE TYPE sos_event_status AS ENUM (
  'REPORTED',
  'DISPATCHED',
  'RESOLVED',
  'CANCELLED'
);

CREATE TYPE sos_event_nature AS ENUM (
  'LABOR',
  'BLEEDING',
  'ACCIDENT',
  'OTHER'
);

CREATE TYPE notification_type AS ENUM (
  'APPOINTMENT',
  'EMERGENCY',
  'GENERAL',
  'HEALTH_TIP',
  'SYSTEM'
);

CREATE TYPE notification_priority AS ENUM (
  'LOW',
  'MEDIUM',
  'HIGH',
  'CRITICAL'
);

CREATE TYPE content_category AS ENUM (
  'GENERAL',
  'NUTRITION',
  'EXERCISE',
  'MENTAL_HEALTH',
  'CHILDCARE',
  'EMERGENCY',
  'POSTPARTUM'
);

CREATE TYPE content_type AS ENUM (
  'ARTICLE',
  'VIDEO',
  'INFOGRAPHIC',
  'AUDIO',
  'QUIZ'
);

-- Create tables
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  phone_number VARCHAR(20) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  role user_role NOT NULL,
  district VARCHAR(255),
  facility_id UUID,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE facilities (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name VARCHAR(255) NOT NULL,
  address TEXT NOT NULL,
  district VARCHAR(255) NOT NULL,
  location GEOGRAPHY(POINT, 4326) NOT NULL,
  facility_type facility_type NOT NULL,
  capacity INTEGER,
  operating_hours JSONB,
  services_offered TEXT[],
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE mothers (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id),
  expected_delivery_date DATE NOT NULL,
  blood_type blood_type,
  health_conditions TEXT[],
  pregnancy_history JSONB,
  risk_level risk_level NOT NULL DEFAULT 'LOW',
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  CONSTRAINT unique_user_id UNIQUE (user_id)
);

CREATE TABLE visits (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  mother_id UUID NOT NULL REFERENCES mothers(id),
  facility_id UUID NOT NULL REFERENCES facilities(id),
  chw_id UUID REFERENCES users(id),
  clinician_id UUID REFERENCES users(id),
  scheduled_time TIMESTAMP WITH TIME ZONE NOT NULL,
  check_in_time TIMESTAMP WITH TIME ZONE,
  check_out_time TIMESTAMP WITH TIME ZONE,
  visit_type visit_type NOT NULL,
  visit_notes TEXT,
  status visit_status NOT NULL DEFAULT 'SCHEDULED',
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE health_metrics (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  mother_id UUID NOT NULL REFERENCES mothers(id),
  visit_id UUID REFERENCES visits(id),
  recorded_by_id UUID REFERENCES users(id),
  recorded_at TIMESTAMP WITH TIME ZONE NOT NULL,
  blood_pressure_systolic FLOAT,
  blood_pressure_diastolic FLOAT,
  fetal_heart_rate FLOAT,
  fetal_movement FLOAT,
  blood_sugar FLOAT,
  hemoglobin_level FLOAT,
  iron_level FLOAT,
  weight FLOAT,
  contractions JSONB,
  notes TEXT,
  is_abnormal BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE sos_events (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  mother_id UUID NOT NULL REFERENCES mothers(id),
  reported_by UUID NOT NULL REFERENCES users(id),
  location GEOGRAPHY(POINT, 4326) NOT NULL,
  nature sos_event_nature NOT NULL,
  description TEXT,
  status sos_event_status NOT NULL DEFAULT 'REPORTED',
  ambulance_id UUID,
  facility_id UUID REFERENCES facilities(id),
  priority INTEGER NOT NULL DEFAULT 1,
  eta TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE notifications (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id),
  title VARCHAR(255) NOT NULL,
  message TEXT NOT NULL,
  type notification_type NOT NULL,
  priority notification_priority NOT NULL DEFAULT 'MEDIUM',
  is_read BOOLEAN NOT NULL DEFAULT FALSE,
  read_at TIMESTAMP WITH TIME ZONE,
  related_entity_id UUID,
  related_entity_type VARCHAR(50),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE education_content (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  title VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  content TEXT NOT NULL,
  category content_category NOT NULL,
  tags TEXT[],
  trimester INTEGER CHECK (trimester BETWEEN 1 AND 3),
  content_type content_type NOT NULL,
  media_url TEXT,
  thumbnail_url TEXT,
  author VARCHAR(255) NOT NULL,
  view_count INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for faster queries
CREATE INDEX idx_facilities_location ON facilities USING GIST (location);
CREATE INDEX idx_facilities_district ON facilities (district);
CREATE INDEX idx_facilities_type ON facilities (facility_type);

CREATE INDEX idx_mothers_user_id ON mothers (user_id);
CREATE INDEX idx_mothers_delivery_date ON mothers (expected_delivery_date);
CREATE INDEX idx_mothers_risk_level ON mothers (risk_level);

CREATE INDEX idx_visits_mother_id ON visits (mother_id);
CREATE INDEX idx_visits_facility_id ON visits (facility_id);
CREATE INDEX idx_visits_scheduled_time ON visits (scheduled_time);
CREATE INDEX idx_visits_status ON visits (status);
CREATE INDEX idx_visits_chw_id ON visits (chw_id);
CREATE INDEX idx_visits_clinician_id ON visits (clinician_id);

CREATE INDEX idx_health_metrics_mother_id ON health_metrics (mother_id);
CREATE INDEX idx_health_metrics_visit_id ON health_metrics (visit_id);
CREATE INDEX idx_health_metrics_recorded_at ON health_metrics (recorded_at);
CREATE INDEX idx_health_metrics_abnormal ON health_metrics (is_abnormal);

CREATE INDEX idx_sos_events_mother_id ON sos_events (mother_id);
CREATE INDEX idx_sos_events_status ON sos_events (status);
CREATE INDEX idx_sos_events_location ON sos_events USING GIST (location);
CREATE INDEX idx_sos_events_priority ON sos_events (priority);

CREATE INDEX idx_notifications_user_id ON notifications (user_id);
CREATE INDEX idx_notifications_read ON notifications (is_read);
CREATE INDEX idx_notifications_type ON notifications (type);
CREATE INDEX idx_notifications_priority ON notifications (priority);

CREATE INDEX idx_education_content_category ON education_content (category);
CREATE INDEX idx_education_content_trimester ON education_content (trimester);
CREATE INDEX idx_education_content_type ON education_content (content_type);
CREATE INDEX idx_education_content_view_count ON education_content (view_count);
