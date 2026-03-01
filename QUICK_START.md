# Quick Start Guide - Logistics Matching Platform

## Setup and Run

### Option 1: Docker Compose (Recommended)
```bash
# Clone the repository
cd /sessions/wizardly-vibrant-ritchie/mnt/google-maps-projects/delivery-api

# Create .env file (copy from example)
cp .env.example .env

# Update .env with your settings (especially JWT_SECRET)
# Edit JWT_SECRET=your_secret_key_here to something strong

# Run with Docker Compose
docker-compose up

# API will be available at http://localhost:8080
```

### Option 2: Local Development
```bash
# Prerequisites: Go 1.21+, MySQL 8.0+

# Set environment variables
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=delivery_user
export DB_PASSWORD=delivery_pass
export DB_NAME=delivery_db
export PORT=8080
export JWT_SECRET=your_secret_key

# Create database
mysql -u root -p -e "CREATE DATABASE delivery_db; CREATE USER 'delivery_user'@'localhost' IDENTIFIED BY 'delivery_pass'; GRANT ALL PRIVILEGES ON delivery_db.* TO 'delivery_user'@'localhost';"

# Install dependencies
go mod download

# Run server
go run cmd/server/main.go
```

## Quick Test Flow

### 1. Register as Driver
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "driver@example.com",
    "password": "securepass123",
    "name": "Taro Yamada",
    "role": "driver",
    "company": "Yamada Logistics",
    "phone": "+81-90-1234-5678"
  }'
```
Save the returned JWT token as `$DRIVER_TOKEN`

### 2. Register as Shipper
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "shipper@example.com",
    "password": "securepass123",
    "name": "Hanako Suzuki",
    "role": "shipper",
    "company": "Suzuki Trading",
    "phone": "+81-90-8765-4321"
  }'
```
Save the returned JWT token as `$SHIPPER_TOKEN`

### 3. Driver Creates a Trip (Tokyo → Kochi)
```bash
curl -X POST http://localhost:8080/api/v1/trips \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $DRIVER_TOKEN" \
  -d '{
    "origin_address": "東京駅, 東京都千代田区",
    "origin_lat": 35.6762,
    "origin_lng": 139.7674,
    "destination_address": "高知駅, 高知県高知市",
    "destination_lat": 33.5553,
    "destination_lng": 133.5307,
    "departure_at": "2024-01-15T08:00:00Z",
    "estimated_arrival": "2024-01-15T20:00:00Z",
    "vehicle_type": "2t",
    "available_weight": 1500,
    "price": 50000,
    "note": "高速道路利用。禁煙車。丁寧な積み込みをお願いします。"
  }'
```
Save the returned trip ID as `$TRIP_ID`

### 4. Shipper Searches for Return Trips (Kochi → Tokyo)
```bash
curl -X POST http://localhost:8080/api/v1/trips/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $SHIPPER_TOKEN" \
  -d '{
    "origin_lat": 33.5553,
    "origin_lng": 133.5307,
    "dest_lat": 35.6762,
    "dest_lng": 139.7674,
    "radius_km": 50,
    "date": "2024-01-15"
  }'
```

This will show:
- **normal_matches**: Trips going Kochi → Tokyo (rare, going opposite)
- **return_matches**: Trips going Tokyo → Kochi that can carry cargo back (帰り便シェア)

### 5. Shipper Requests Match
```bash
curl -X POST http://localhost:8080/api/v1/matches \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $SHIPPER_TOKEN" \
  -d '{
    "trip_id": '$TRIP_ID',
    "cargo_weight": 800,
    "cargo_description": "精密機器・常温保管・要温度管理",
    "message": "丁寧な扱いをお願いします。高速道路での運搬をお願いします。"
  }'
```
Save the returned match ID as `$MATCH_ID`

### 6. Driver Approves Match
```bash
curl -X PUT http://localhost:8080/api/v1/matches/$MATCH_ID/approve \
  -H "Authorization: Bearer $DRIVER_TOKEN"
```

### 7. Driver Records Location During Trip
```bash
# Simulate location updates every 10 minutes
curl -X POST http://localhost:8080/api/v1/tracking \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $DRIVER_TOKEN" \
  -d '{
    "trip_id": '$TRIP_ID',
    "lat": 35.6762,
    "lng": 139.7674
  }'

# After an hour, send another location
curl -X POST http://localhost:8080/api/v1/tracking \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $DRIVER_TOKEN" \
  -d '{
    "trip_id": '$TRIP_ID',
    "lat": 34.5,
    "lng": 136.5
  }'
```

### 8. View Tracking History
```bash
curl -X GET http://localhost:8080/api/v1/tracking/$TRIP_ID \
  -H "Authorization: Bearer $DRIVER_TOKEN"
```

### 9. Get Latest Location
```bash
curl -X GET http://localhost:8080/api/v1/tracking/$TRIP_ID/latest \
  -H "Authorization: Bearer $DRIVER_TOKEN"
```

### 10. Complete Match
```bash
curl -X PUT http://localhost:8080/api/v1/matches/$MATCH_ID/complete \
  -H "Authorization: Bearer $DRIVER_TOKEN"
```

### 11. View Admin Dashboard (with admin account)
First, register as admin or have the initial admin user:
```bash
# Assuming you have admin token
curl -X GET http://localhost:8080/api/v1/admin/stats \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

## Key Endpoints by Role

### Driver
- `POST /api/v1/auth/register` - Register
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/trips` - Create trip
- `PUT /api/v1/trips/:id` - Update trip
- `DELETE /api/v1/trips/:id` - Cancel trip
- `PUT /api/v1/matches/:id/approve` - Approve cargo match
- `PUT /api/v1/matches/:id/reject` - Reject cargo match
- `POST /api/v1/tracking` - Report location

### Shipper
- `POST /api/v1/auth/register` - Register
- `POST /api/v1/auth/login` - Login
- `GET /api/v1/trips` - View available trips
- `POST /api/v1/trips/search` - Search for return trips
- `POST /api/v1/matches` - Request cargo match
- `GET /api/v1/tracking/:trip_id/latest` - Track shipment

### Admin
- All user endpoints
- `GET /api/v1/admin/stats` - Dashboard statistics
- `GET /api/v1/admin/users` - List all users
- `GET /api/v1/admin/trips` - List all trips
- `GET /api/v1/admin/matches` - List all matches
- `PUT /api/v1/admin/users/:id/role` - Change user role

## Common Issues & Solutions

### Issue: JWT validation fails
**Solution**: Make sure JWT_SECRET is set and consistent between registration and API requests

### Issue: Database connection refused
**Solution**: Check MySQL is running and credentials in .env match

### Issue: Permission denied on admin endpoints
**Solution**: Ensure user role is "admin" - use `/api/v1/admin/users/:id/role` to change role

### Issue: No return trips found in search
**Solution**:
- Check trip coordinates are correct (use real lat/lng or test coordinates)
- Verify radius_km is large enough
- Confirm trips exist and have status="open"

### Issue: Cargo weight exceeds available weight
**Solution**: Create a trip with larger `available_weight` or request match with less cargo

## Project Structure
```
delivery-api/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point
├── internal/
│   ├── model/                      # Data models
│   ├── repository/                 # Database operations
│   ├── handler/                    # HTTP handlers
│   ├── middleware/                 # Request middleware
│   ├── router/                     # Route definitions
│   └── util/                       # Helper functions
├── go.mod                          # Dependencies
├── docker-compose.yml              # Docker setup
├── .env.example                    # Environment template
└── README.md                       # Project docs
```

## Performance Tips

1. **Search Optimization**: Increase radius_km for wider search area
2. **Tracking**: Send location updates every 10 minutes (balance accuracy vs server load)
3. **Database**: Keep connection pool size reasonable (default fine for most cases)
4. **Caching**: Future: implement Redis caching for frequent searches

## Next Steps

1. Read API_DOCUMENTATION.md for full endpoint reference
2. Read IMPLEMENTATION_NOTES.md for technical architecture
3. Customize JWT_SECRET for production
4. Add rate limiting for production
5. Set up monitoring and logging
6. Test with real location data

## Support

For issues or questions:
1. Check API_DOCUMENTATION.md for endpoint details
2. Review error messages in server logs
3. Verify request format matches examples
4. Check user roles and permissions
