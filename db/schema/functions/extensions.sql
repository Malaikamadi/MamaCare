-- PostgreSQL extensions for MamaCare SL

-- Enable UUID generation for unique identifiers
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable PostGIS for geospatial features (location tracking, distance calculations)
CREATE EXTENSION IF NOT EXISTS "postgis";

-- Enable trigram support for text search capabilities
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Enable cryptographic functions for secure data handling
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
