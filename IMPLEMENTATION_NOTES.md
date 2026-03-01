# Implementation Notes - Logistics Matching Platform

## Summary of Changes

This document provides technical implementation details for the logistics matching platform expansion.

## Files Created

### Models (5 new files)
1. **internal/model/user.go** - User accounts with roles (driver, shipper, admin)
2. **internal/model/vehicle.go** - Vehicle information for drivers
3. **internal/model/trip.go** - Trip/route definitions with location and capacity info
4. **internal/model/match.go** - Cargo-to-trip matches with approval workflow
5. **internal/model/tracking.go** - Real-time location tracking for trips

### Repositories (4 new files)
1. **internal/repository/user.go** - CRUD operations for users
   - Methods: GetAll, GetByID, GetByEmail, Create, Update, Delete
   - Email is unique indexed for fast lookups

2. **internal/repository/trip.go** - Trip database operations
   - Methods: GetAll, GetByID, GetByDriverID, GetOpenTrips, GetTripsAfterDate, Create, Update, Delete, GetTripsByOriginDestination
   - Preloads Driver relationship

3. **internal/repository/match.go** - Match database operations
   - Methods: GetAll, GetByID, GetByTripID, GetByShipperID, GetPendingMatches, Create, Update, Delete
   - Preloads Trip and Shipper relationships

4. **internal/repository/tracking.go** - Tracking record operations
   - Methods: GetByTripID, GetLatestByTripID, Create, Delete
   - Records GPS coordinates with timestamps

### Handlers (4 new files + 1 updated)
1. **internal/handler/auth.go** - Authentication endpoints
   - Register: Creates user with bcrypt hashed password
   - Login: Validates credentials and returns JWT token
   - Supports driver/shipper roles

2. **internal/handler/trip.go** - Trip management
   - CRUD operations for trips
   - **SearchTrips**: Special endpoint for return trip matching
     - Searches for trips where:
       - Normal: origin near search_origin AND destination near search_destination
       - Return (帰り便): origin near search_destination AND destination near search_origin
     - Uses Haversine formula with configurable radius
     - Filters by optional date

3. **internal/handler/match.go** - Match workflow
   - CreateMatch: Shipper requests match with cargo details
   - ApproveMatch: Driver approves match request
   - RejectMatch: Driver rejects match request
   - CompleteMatch: Mark match as completed
   - Validates cargo weight against trip capacity

4. **internal/handler/tracking.go** - Location tracking
   - RecordLocation: Records GPS point for a trip
   - GetTrackingHistory: Retrieves all location points for a trip
   - GetLatestLocation: Gets most recent location for real-time map display

5. **internal/handler/admin.go** - Admin dashboard
   - GetStats: Platform-wide statistics (users, trips, matches, active trips, etc.)
   - GetUsers, GetTrips, GetMatches: List all data
   - UpdateUserRole: Change user roles (driver → admin, etc.)

### Middleware (1 new file)
1. **internal/middleware/auth.go** - Authentication & authorization
   - AuthMiddleware: Validates JWT from Authorization header
   - AdminMiddleware: Checks for admin role
   - DriverMiddleware: Checks for driver role
   - ShipperMiddleware: Checks for shipper role
   - Sets user_id, email, and role in gin context for downstream handlers

### Utilities (2 new files)
1. **internal/util/haversine.go** - Distance calculations
   - CalculateDistance: Great circle distance between two lat/lng points
   - IsWithinRadius: Checks if point is within radius of another
   - Uses standard Haversine formula (Earth radius: 6371 km)
   - Used for return trip search radius matching

2. **internal/util/jwt.go** - JWT token management
   - Claims struct: Contains UserID, Email, Role
   - GenerateToken: Creates signed JWT with 24-hour expiration
   - ParseToken: Validates and extracts claims from token
   - Uses HS256 signing algorithm

### Updated Files (3 files)
1. **internal/router/router.go** - Route definitions
   - Added auth routes (public)
   - Protected routes with AuthMiddleware
   - Role-specific middleware for endpoints
   - Admin routes under separate group with AdminMiddleware
   - Initializes all new repositories and handlers
   - Reads JWT_SECRET from environment

2. **cmd/server/main.go** - Server initialization
   - Added AutoMigrate for all new models
   - Maintains backward compatibility with Delivery model

3. **go.mod** - Dependencies
   - Added: github.com/golang-jwt/jwt/v5 v5.2.0
   - Added: golang.org/x/crypto v0.17.0 (for bcrypt)

### Configuration Files
1. **.env.example** - Added JWT_SECRET variable
2. **docker-compose.yml** - Added JWT_SECRET environment configuration
3. **API_DOCUMENTATION.md** - Complete API reference (NEW)
4. **IMPLEMENTATION_NOTES.md** - This file (NEW)

## Key Design Decisions

### 1. Authentication Strategy
- **Password Hashing**: bcrypt with default cost (12 rounds)
- **Token Format**: JWT with HS256 signature
- **Token Expiration**: 24 hours for security
- **Claims**: UserID, Email, Role for authorization decisions

### 2. Return Trip Matching (帰り便シェア)
The core feature uses a geometric approach:
- Origin: A location (lat/lng)
- Destination: B location (lat/lng)
- Shipper searches: A → B

The system returns:
- **Normal trips**: Trips where origin ≈ A AND destination ≈ B
- **Return trips**: Trips where origin ≈ B AND destination ≈ A
  - These are drivers going back (帰り便) and can carry cargo return

Search radius defaults to 50km, uses Haversine formula for accuracy.

### 3. Match Workflow
```
1. Driver posts Trip with available_weight
2. Shipper creates Match request with cargo_weight
3. Trip status: open → matched (when approved)
4. Shipper waits for approval
5. Driver approves/rejects Match
6. If approved: Match status pending → approved
7. Upon delivery: Match status → completed
```

### 4. Repository Pattern
- Each model has dedicated repository
- GORM for ORM with preloading for relationships
- Consistent CRUD interface
- Error handling with meaningful messages

### 5. Role-Based Access Control
```
Public:
  - POST /auth/register
  - POST /auth/login

Authenticated (all roles):
  - GET /trips
  - GET /trips/:id
  - POST /tracking
  - GET /tracking/*
  - GET /matches

Driver Only:
  - POST /trips (create)
  - PUT /matches/:id/approve
  - PUT /matches/:id/reject

Shipper Only:
  - POST /matches (create)

Admin Only:
  - GET /admin/stats
  - GET /admin/users
  - GET /admin/trips
  - GET /admin/matches
  - PUT /admin/users/:id/role
```

### 6. Database Design
- All models use GORM with MySQL compatibility
- Foreign keys for relationships (Driver in Trip, Trip/Shipper in Match)
- Preload relationships to avoid N+1 queries
- Email unique index for fast user lookups
- Timestamps (created_at, updated_at) on all models

### 7. Backward Compatibility
- Legacy Delivery model untouched
- Legacy delivery endpoints still functional
- New system runs parallel to old one
- Can migrate data gradually if needed

## Technical Stack

- **Framework**: Gin Web Framework
- **ORM**: GORM with MySQL driver
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **Password Hashing**: bcrypt (golang.org/x/crypto)
- **Language**: Go 1.21+
- **Database**: MySQL 8.0

## Testing Recommendations

### Unit Tests Needed
- Haversine distance calculations (edge cases)
- JWT token generation and validation
- Password hashing and comparison
- Repository CRUD operations

### Integration Tests Needed
- Auth flow (register → login)
- Trip creation and search
- Match workflow (create → approve → complete)
- Tracking location recording
- Admin statistics calculation

### Load Testing Recommended
- Concurrent trip searches (Haversine is CPU-intensive)
- Location tracking writes during active trips
- JWT validation under high request volume

## Security Considerations

1. **Password Storage**: Bcrypt with cost 12
2. **Token Security**: HS256 with strong secret (change in production)
3. **JWT Secret**: Must be environment variable, not hardcoded
4. **CORS**: Enabled for cross-origin requests
5. **Input Validation**: Binding validation on all request bodies
6. **Email Uniqueness**: Enforced at database level

## Performance Optimizations

1. **Database**:
   - Email unique index for O(1) user lookups
   - Preload relationships to avoid N+1 queries
   - Trip status filtering in GetOpenTrips

2. **API**:
   - JWT cached in gin context
   - Direct coordinate search without geocoding

3. **Distance Calculation**:
   - Haversine formula is O(1) per comparison
   - No database calls needed for search

## Future Enhancements

1. **Caching**:
   - Cache open trips in Redis
   - Cache search results for 5 minutes
   - User profile caching

2. **Real-Time Features**:
   - WebSocket for live tracking
   - Push notifications on match approval
   - Live chat between driver and shipper

3. **Ratings & Reviews**:
   - Driver ratings by shippers
   - Shipper reliability scores
   - Trip history and statistics

4. **Payment Integration**:
   - Stripe/PayPal integration
   - Escrow for cargo safety
   - Dispute resolution

5. **Advanced Matching**:
   - Machine learning for optimal matches
   - Price negotiation workflow
   - Bulk shipment consolidation

6. **Insurance & Compliance**:
   - Cargo insurance options
   - Compliance tracking
   - Document uploads

## Deployment Checklist

- [ ] Set JWT_SECRET to strong random value
- [ ] Configure database credentials in .env
- [ ] Use HTTPS in production
- [ ] Set Gin mode to release (release mode)
- [ ] Enable database backups
- [ ] Monitor error logs
- [ ] Set up load balancer if scaling
- [ ] Configure CORS for production domain
- [ ] Run database migrations before deployment
- [ ] Test auth flow end-to-end
- [ ] Test trip search with real coordinates

## Known Limitations

1. No pagination on list endpoints (add in future)
2. No filtering parameters on most GET endpoints
3. Haversine formula assumes spherical Earth (accurate to ~0.5%)
4. No transaction management for match approval workflow
5. No automatic match expiration (trips stay open indefinitely)
6. Admin endpoints unprotected from role changes during requests

## Code Statistics

- **New Models**: 5 files
- **New Repositories**: 4 files
- **New Handlers**: 4 files
- **New Middleware**: 1 file
- **New Utils**: 2 files
- **Updated Files**: 3 files
- **Total New Lines**: ~2500
- **All files follow Go best practices and naming conventions**
