# MamaCare Go Services

This repository contains the Go microservices for the MamaCare maternal health platform.

## Services

### Authentication Service

The authentication service handles user authentication and authorization using Firebase Phone Authentication and JWTs.

#### Features

- Firebase Phone Authentication integration
- JWT token generation and validation
- Role-based access control
- Hasura authentication webhook for GraphQL permissions

#### Configuration

Create a `configs/auth.yaml` file with the following configuration:

```yaml
# Server configuration
server:
  address: ":8080"
  read_timeout: 5s
  write_timeout: 10s
  idle_timeout: 120s
  shutdown_timeout: 10s

# Database configuration
database:
  host: localhost
  port: 5432
  user: postgres
  password: your-password-here  # Use env var in production
  name: mamacare
  sslmode: disable    # Use 'require' in production

# Authentication configuration
auth:
  # JWT configuration
  jwt_secret: "your-secret-key-replace-in-prod"  # Use env var in production
  jwt_expiry: 24h
  hasura_namespace: "https://hasura.io/jwt/claims"

# Firebase configuration
firebase:
  credentials_file: "./firebase-credentials.json"  # Path to your Firebase credentials file
  project_id: "mamacare-app"  # Your Firebase project ID
```

#### Firebase Setup

1. Create a Firebase project at [firebase.google.com](https://firebase.google.com)
2. Enable Phone Authentication in the Authentication section
3. Generate a service account key from Project Settings > Service accounts
4. Save the JSON key as `firebase-credentials.json` in the same directory as the auth service binary

#### Running the Service

```bash
# Build the service
cd cmd/auth
go build

# Run the service
./auth
```

## Development

### Project Structure

The project follows clean architecture principles:

- `cmd/` - Service entry points
- `internal/` - Application code not meant to be imported by other projects
  - `domain/` - Domain models and repository interfaces
  - `app/` - Application business logic
  - `infra/` - Infrastructure implementations (database, HTTP, Firebase)
  - `port/` - Input/output adapters (HTTP handlers, middleware)
- `pkg/` - Shared packages that can be imported by other projects

### Tech Stack

- Go 1.22
- PostgreSQL 14 with PostGIS for geospatial queries
- Firebase Admin SDK for authentication
- Chi router for HTTP handling
- Zerolog for structured logging

## API Documentation

### Authentication Endpoints

#### POST /auth/login

Authenticates a user with phone verification.

Request:
```json
{
  "phone_number": "+123456789",
  "verification_code": "123456"
}
```

Response:
```json
{
  "user": {
    "id": "user-uuid",
    "name": "User Name",
    "email": "user@example.com",
    "role": "MOTHER"
  },
  "token": "firebase-auth-token"
}
```

#### POST /auth/refresh

Refreshes an authentication token.

Request: Include Bearer token in Authorization header

Response:
```json
{
  "token": "new-jwt-token",
  "expires_at": 1620000000
}
```

#### GET /user/hasura-jwt

Generates a Hasura-specific JWT for the authenticated user.

Request: Include Bearer token in Authorization header

Response:
```json
{
  "token": "hasura-jwt-token",
  "expires_at": 1620000000
}
```

#### POST /hasura-webhook

Webhook for Hasura authentication.

Request: Include Bearer token in Authorization header

Response:
```json
{
  "X-Hasura-Role": "mother",
  "X-Hasura-User-Id": "user-uuid",
  "claims": {
    "x-hasura-allowed-roles": ["mother"],
    "x-hasura-default-role": "mother",
    "x-hasura-user-id": "user-uuid",
    "x-hasura-mother-id": "mother-uuid"
  }
}
```
