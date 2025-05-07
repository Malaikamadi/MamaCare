-- Data validation domains for MamaCare SL

-- Phone numbers (Sierra Leone format)
CREATE DOMAIN phone_number AS TEXT
CHECK (
  VALUE ~ '^\+?232[0-9]{9}$' -- Sierra Leone country code + 9 digits
);

-- Email validation
CREATE DOMAIN email AS TEXT
CHECK (
  VALUE ~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
);

-- GPS coordinates validation
CREATE DOMAIN latitude AS DECIMAL
CHECK (VALUE BETWEEN -90 AND 90);

CREATE DOMAIN longitude AS DECIMAL
CHECK (VALUE BETWEEN -180 AND 180);

-- Simple text validation for required fields (non-empty)
CREATE DOMAIN non_empty_text AS TEXT
CHECK (
  LENGTH(TRIM(VALUE)) > 0
);

-- Date validation for future dates (for appointments)
CREATE DOMAIN future_date AS DATE
CHECK (
  VALUE >= CURRENT_DATE
);

-- Password strength validation
CREATE DOMAIN strong_password AS TEXT
CHECK (
  LENGTH(VALUE) >= 8 AND  -- At least 8 characters
  VALUE ~ '[A-Z]' AND     -- At least one uppercase letter
  VALUE ~ '[a-z]' AND     -- At least one lowercase letter
  VALUE ~ '[0-9]'         -- At least one number
);
