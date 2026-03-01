# Logistics Matching Platform API - 帰り便シェア (Return Trip Sharing)

## Overview

This API has been expanded into a full logistics matching platform that enables drivers and shippers to collaborate on shared deliveries, with special focus on "帰り便シェア" (return trip sharing) - matching cargo with empty truck space on return journeys.

## Architecture

### Models

#### User (`internal/model/user.go`)
- Represents platform users (drivers, shippers, admins)
- Fields: ID, Email, PasswordHash, Name, Role, Company, Phone, CreatedAt, UpdatedAt
- Roles: `driver`, `shipper`, `admin`

#### Vehicle (`internal/model/vehicle.go`)
- Represents a driver's vehicle
- Fields: ID, UserID, Type, MaxWeight, PlateNumber, CreatedAt, UpdatedAt
- Types: 軽トラ, 2t, 4t, 10t, 大型

#### Trip (`internal/model/trip.go`)
- Represents a driver's planned trip/route
- Fields: ID, DriverID, OriginAddress, OriginLat, OriginLng, DestinationAddress, DestinationLat, DestinationLng, DepartureAt, EstimatedArrival, VehicleType, AvailableWeight, Price, Status, Note, DelayMinutes, CreatedAt, UpdatedAt
- Status: `open`, `matched`, `in_transit`, `completed`, `cancelled`

#### Match (`internal/model/match.go`)
- Represents a match between a trip and cargo shipment
- Fields: ID, TripID, ShipperID, CargoWeight, CargoDescription, Status, Message, CreatedAt, UpdatedAt
- Status: `pending`, `approved`, `rejected`, `completed`

#### Tracking (`internal/model/tracking.go`)
- Real-time location tracking for trips
- Fields: ID, TripID, Lat, Lng, RecordedAt

#### Delivery (`internal/model/delivery.go`)
- Legacy delivery model (preserved from original implementation)

## API Endpoints

### Authentication (Public)

#### POST `/api/v1/auth/register`
Register a new user
```json
Request:
{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe",
  "role": "driver|shipper",
  "company": "Company Name",
  "phone": "+81-90-xxxx-xxxx"
}

Response (201):
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "role": "driver",
    "company": "Company Name",
    "phone": "+81-90-xxxx-xxxx",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### POST `/api/v1/auth/login`
Login user and receive JWT token
```json
Request:
{
  "email": "user@example.com",
  "password": "password123"
}

Response (200):
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": { ... }
}
```

### Trips (Protected - JWT Required)

#### GET `/api/v1/trips`
Get all trips
```
Response (200): Array of Trip objects
```

#### GET `/api/v1/trips/:id`
Get trip by ID
```
Response (200): Trip object
```

#### POST `/api/v1/trips` (Driver Only)
Create new trip
```json
Request:
{
  "origin_address": "東京駅",
  "origin_lat": 35.6762,
  "origin_lng": 139.7674,
  "destination_address": "高知駅",
  "destination_lat": 33.5553,
  "destination_lng": 133.5307,
  "departure_at": "2024-01-15T08:00:00Z",
  "estimated_arrival": "2024-01-15T16:00:00Z",
  "vehicle_type": "2t",
  "available_weight": 1500,
  "price": 50000,
  "note": "高速道路利用"
}

Response (201): Trip object with status="open"
```

#### PUT `/api/v1/trips/:id`
Update trip
```
Response (200): Updated Trip object
```

#### DELETE `/api/v1/trips/:id`
Delete trip
```
Response (200): {"message": "trip deleted successfully"}
```

#### POST `/api/v1/trips/search`
Search for return trips (帰り便シェア)
```json
Request:
{
  "origin_lat": 35.6762,
  "origin_lng": 139.7674,
  "dest_lat": 33.5553,
  "dest_lng": 133.5307,
  "radius_km": 50,
  "date": "2024-01-15"
}

Response (200):
{
  "normal_matches": [...],
  "return_matches": [...],
  "total_matches": 5,
  "normal_count": 2,
  "return_count": 3
}
```

**Special Return Trip Matching:**
- Searches for trips where origin_lat/lng ≈ search dest_lat/lng AND dest_lat/lng ≈ search origin_lat/lng
- These are drivers returning (帰り便) who can transport cargo back to origin
- Uses Haversine formula with configurable radius (default 50km)

### Matches (Protected - JWT Required)

#### GET `/api/v1/matches`
Get all matches
```
Response (200): Array of Match objects
```

#### GET `/api/v1/matches/:id`
Get match by ID
```
Response (200): Match object
```

#### POST `/api/v1/matches` (Shipper Only)
Request a match for cargo
```json
Request:
{
  "trip_id": 1,
  "cargo_weight": 800,
  "cargo_description": "精密機器 / 常温保管",
  "message": "丁寧な扱いをお願いします"
}

Response (201): Match object with status="pending"
```

#### PUT `/api/v1/matches/:id/approve` (Driver Only)
Driver approves match request
```
Response (200): Match object with status="approved"
```

#### PUT `/api/v1/matches/:id/reject` (Driver Only)
Driver rejects match request
```
Response (200): Match object with status="rejected"
```

#### PUT `/api/v1/matches/:id/complete`
Mark match as completed
```
Response (200): Match object with status="completed"
```

### Tracking (Protected - JWT Required)

#### POST `/api/v1/tracking`
Record location (driver sends ~every 10 minutes)
```json
Request:
{
  "trip_id": 1,
  "lat": 35.6762,
  "lng": 139.7674
}

Response (201): Tracking object
```

#### GET `/api/v1/tracking/:trip_id`
Get tracking history for trip
```
Response (200): Array of Tracking objects (ordered by recorded_at)
```

#### GET `/api/v1/tracking/:trip_id/latest`
Get latest location for trip
```
Response (200): Latest Tracking object
```

### Admin (Protected - Admin Only)

#### GET `/api/v1/admin/stats`
Get dashboard statistics
```
Response (200):
{
  "total_users": 150,
  "total_trips": 320,
  "total_matches": 240,
  "active_trips": 15,
  "completed_trips": 180,
  "pending_matches": 20,
  "approved_matches": 150
}
```

#### GET `/api/v1/admin/users`
Get all users
```
Response (200): Array of User objects
```

#### GET `/api/v1/admin/trips`
Get all trips
```
Response (200): Array of Trip objects
```

#### GET `/api/v1/admin/matches`
Get all matches
```
Response (200): Array of Match objects
```

#### PUT `/api/v1/admin/users/:id/role`
Change user role
```json
Request:
{
  "role": "driver|shipper|admin"
}

Response (200): {"message": "user role updated successfully"}
```

### Delivery (Legacy - Public)

#### GET `/api/v1/deliveries`
Get all deliveries
```
Response (200): Array of Delivery objects
```

#### GET `/api/v1/deliveries/:id`
Get delivery by ID
```
Response (200): Delivery object
```

#### POST `/api/v1/deliveries`
Create new delivery
```json
Request:
{
  "name": "Delivery Name",
  "address": "Address",
  "lat": 35.6762,
  "lng": 139.7674,
  "note": "Note"
}

Response (201): Delivery object
```

#### PUT `/api/v1/deliveries/:id`
Update delivery
```
Response (200): Updated Delivery object
```

#### DELETE `/api/v1/deliveries/:id`
Delete delivery
```
Response (200): {"message": "Delivery deleted successfully"}
```

## Authentication

All protected endpoints require JWT token in Authorization header:
```
Authorization: Bearer <JWT_TOKEN>
```

### JWT Token Details
- Algorithm: HS256
- Expiration: 24 hours
- Claims: UserID, Email, Role
- Secret: JWT_SECRET environment variable

## File Structure

```
internal/
├── model/
│   ├── delivery.go      (legacy)
│   ├── user.go          (NEW)
│   ├── vehicle.go       (NEW)
│   ├── trip.go          (NEW)
│   ├── match.go         (NEW)
│   └── tracking.go      (NEW)
├── repository/
│   ├── delivery.go      (existing)
│   ├── user.go          (NEW)
│   ├── trip.go          (NEW)
│   ├── match.go         (NEW)
│   └── tracking.go      (NEW)
├── handler/
│   ├── delivery.go      (existing)
│   ├── auth.go          (NEW)
│   ├── trip.go          (NEW)
│   ├── match.go         (NEW)
│   ├── tracking.go      (NEW)
│   └── admin.go         (NEW)
├── middleware/
│   ├── cors.go          (existing)
│   └── auth.go          (NEW)
├── router/
│   └── router.go        (UPDATED)
└── util/
    ├── haversine.go     (NEW)
    └── jwt.go           (NEW)

cmd/server/
└── main.go              (UPDATED)
```

## Dependencies Added

```
github.com/golang-jwt/jwt/v5 v5.2.0      # JWT token handling
golang.org/x/crypto v0.17.0               # bcrypt password hashing
```

## Environment Variables

```
DB_HOST=localhost
DB_PORT=3306
DB_USER=delivery_user
DB_PASSWORD=delivery_pass
DB_NAME=delivery_db
PORT=8080
GOOGLE_MAPS_API_KEY=your_api_key_here
JWT_SECRET=your_secret_key_here           # (NEW)
```

## Database Schema

All models are automatically migrated on startup via `db.AutoMigrate()`:
- users (email is unique)
- vehicles
- trips
- matches
- trackings
- deliveries (legacy)

## Key Features

### 1. Return Trip Matching (帰り便シェア)
- Drivers can post empty truck capacity for return journeys
- Shippers can search for trips going in opposite direction
- Haversine distance calculation for radius-based matching
- Default search radius: 50km (configurable)

### 2. Real-Time Tracking
- Location updates recorded for active trips
- Latest location retrieval for map display
- Full tracking history available

### 3. User Role Management
- Driver: Can create trips, approve matches, track location
- Shipper: Can request matches, view trips
- Admin: Can view all data, manage user roles, dashboard stats

### 4. JWT Authentication
- Secure token-based authentication
- 24-hour token expiration
- Role-based access control via middleware

### 5. Backward Compatibility
- Legacy delivery endpoints preserved
- Existing delivery model unchanged
- Can run alongside new logistics system

## Running the Application

### With Docker Compose
```bash
docker-compose up
```

### Manual Setup
```bash
# Install dependencies
go mod download

# Set environment variables
export JWT_SECRET="your_secret_key"
export DB_HOST="localhost"
export DB_PORT="3306"
export DB_USER="delivery_user"
export DB_PASSWORD="delivery_pass"
export DB_NAME="delivery_db"

# Run server
go run cmd/server/main.go
```

## Testing the API

### Register User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "driver@example.com",
    "password": "password123",
    "name": "John Driver",
    "role": "driver",
    "company": "Driver Co",
    "phone": "+81-90-xxxx-xxxx"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "driver@example.com",
    "password": "password123"
  }'
```

### Create Trip (with JWT token)
```bash
curl -X POST http://localhost:8080/api/v1/trips \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{
    "origin_address": "東京駅",
    "origin_lat": 35.6762,
    "origin_lng": 139.7674,
    "destination_address": "高知駅",
    "destination_lat": 33.5553,
    "destination_lng": 133.5307,
    "departure_at": "2024-01-15T08:00:00Z",
    "estimated_arrival": "2024-01-15T16:00:00Z",
    "vehicle_type": "2t",
    "available_weight": 1500,
    "price": 50000
  }'
```

### Search Return Trips
```bash
curl -X POST http://localhost:8080/api/v1/trips/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{
    "origin_lat": 35.6762,
    "origin_lng": 139.7674,
    "dest_lat": 33.5553,
    "dest_lng": 133.5307,
    "radius_km": 50,
    "date": "2024-01-15"
  }'
```

## Code Quality

All files follow Go best practices:
- Clear separation of concerns (model, repository, handler)
- Dependency injection via constructors
- Error handling with appropriate HTTP status codes
- Goroutine-safe concurrent access via GORM
- Type-safe with Go's strong typing system
