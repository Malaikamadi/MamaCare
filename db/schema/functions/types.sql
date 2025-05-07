-- Custom enum types for MamaCare SL

-- 1. Risk assessment results (as mentioned in blueprint)
CREATE TYPE risk_level AS ENUM (
  'GREEN',    -- Low risk
  'YELLOW',   -- Medium risk
  'RED'       -- High risk/emergency
);

-- 2. User roles
CREATE TYPE user_role AS ENUM (
  'MOTHER',   -- Pregnant/postpartum mother
  'CHW',      -- Community Health Worker
  'CLINICIAN',-- Hospital/clinic staff  
  'ADMIN'     -- System administrator
);

-- 3. Visit types
CREATE TYPE visit_type AS ENUM (
  'ANTENATAL',    -- Pregnancy checkup
  'POSTNATAL',    -- After delivery checkup
  'VACCINATION',  -- Child vaccination
  'GENERAL',      -- General health check
  'EMERGENCY'     -- Emergency visit
);

-- 4. Visit status
CREATE TYPE visit_status AS ENUM (
  'SCHEDULED',    -- Upcoming visit
  'COMPLETED',    -- Visit happened
  'MISSED',       -- Visit didn't happen
  'RESCHEDULED'   -- Visit changed date
);

-- 5. SOS emergency request status
CREATE TYPE sos_status AS ENUM (
  'OPEN',       -- Request created, awaiting response
  'ACCEPTED',   -- A facility/CHW has taken the case
  'CLOSED'      -- Request resolved/finished
);

-- 6. Device platforms for notification management
CREATE TYPE device_platform AS ENUM (
  'IOS',
  'ANDROID',
  'WEB'
);

-- 7. Pregnancy stage tracking
CREATE TYPE pregnancy_stage AS ENUM (
  'FIRST_TRIMESTER',
  'SECOND_TRIMESTER',
  'THIRD_TRIMESTER',
  'POSTPARTUM'
);

-- 8. Types of notifications to send
CREATE TYPE notification_type AS ENUM (
  'VISIT_REMINDER',      -- Upcoming appointment reminder
  'VACCINE_DUE',         -- Vaccine is due soon
  'DANGER_SIGN_ALERT',   -- Response to reported danger sign
  'SOS_RESPONSE',        -- Response to SOS request
  'EDUCATIONAL'          -- Educational content push
);

-- 9. Child growth tracking categories
CREATE TYPE growth_status AS ENUM (
  'NORMAL',
  'UNDERWEIGHT',
  'OVERWEIGHT',
  'STUNTED',
  'WASTED'
);

-- 10. Language preferences for content and notifications
CREATE TYPE language_preference AS ENUM (
  'ENGLISH',
  'KRIO',
  'MENDE',
  'TEMNE'
);

-- 11. Ambulance/emergency vehicle status
CREATE TYPE ambulance_status AS ENUM (
  'AVAILABLE',     -- Ready for dispatch
  'DISPATCHED',    -- En route to patient
  'TRANSPORTING',  -- Patient on board, en route to facility
  'RETURNING',     -- Returning to base/station
  'MAINTENANCE'    -- Temporarily unavailable
);
