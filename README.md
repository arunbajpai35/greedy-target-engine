# Targeting Engine - GreedyGame Backend Assignment

A high-performance targeting engine microservice that routes campaigns to requests based on targeting criteria. Built with Go, PostgreSQL, and Docker.

## 🚀 Features

- **High Performance**: Optimized for read-heavy workloads with billions of delivery requests
- **Scalable Architecture**: Horizontal and vertical scaling support
- **Complex Targeting Rules**: Support for include/exclude rules across multiple dimensions
- **Real-time Updates**: Reacts to database changes automatically
- **Comprehensive Testing**: Unit and integration tests with high coverage
- **Production Ready**: Graceful shutdown, health checks, and proper error handling

## 🏗️ Architecture

### Core Entities

1. **Campaign**: Central entity representing an advertisement
   - `cid`: Unique campaign identifier
   - `name`: Campaign name
   - `img`: Image creative URL
   - `cta`: Call to action text
   - `status`: ACTIVE or INACTIVE

2. **Targeting Rule**: Defines where campaigns can run
   - Include/Exclude rules for Country, OS, and App ID
   - Support for multiple values per dimension
   - Case-insensitive matching

3. **Delivery**: Service that matches requests to campaigns
   - Accepts app, country, and OS parameters
   - Returns matching campaigns or 204 for no matches

### Database Design

- **campaigns**: Stores campaign information
- **targeting_rules**: Stores targeting criteria with array support
- **Indexes**: Optimized for read-heavy workloads

## 🛠️ Setup & Installation

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)

### Quick Start with Docker

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd greedy-target-engine
   ```

2. **Start the services**
   ```bash
   docker-compose up -d
   ```

3. **Verify the setup**
   ```bash
   # Check health endpoint
   curl http://localhost:8080/healthz
   
   # Test delivery endpoint
   curl "http://localhost:8080/v1/delivery?app=com.gametion.ludokinggame&country=us&os=android"
   ```

### Local Development

1. **Start PostgreSQL**
   ```bash
   docker-compose up postgres -d
   ```

2. **Run migrations**
   ```bash
   # The migrations run automatically when PostgreSQL starts
   # Or manually:
   docker exec -i $(docker-compose ps -q postgres) psql -U postgres -d targeting_db < db/migrations/init.sql
   docker exec -i $(docker-compose ps -q postgres) psql -U postgres -d targeting_db < db/migrations/seed.sql
   ```

3. **Run the application**
   ```bash
   go run cmd/server/main.go
   ```

## 📡 API Documentation

### Health Check

```http
GET /healthz
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Delivery Endpoint

```http
GET /v1/delivery?app={app_id}&country={country}&os={operating_system}
```

**Parameters:**
- `app` (required): Application identifier (e.g., "com.gametion.ludokinggame")
- `country` (required): Country code (e.g., "us", "germany")
- `os` (required): Operating system (e.g., "android", "ios", "web")

**Responses:**

**Success (200 OK):**
```json
[
  {
    "cid": "spotify",
    "name": "Spotify - Music for everyone",
    "img": "https://somelink",
    "cta": "Download"
  },
  {
    "cid": "subwaysurfer",
    "name": "Subway Surfer",
    "img": "https://somelink3",
    "cta": "Play"
  }
]
```

**No Matches (204 No Content):**
```http
HTTP/1.1 204 No Content
```

**Bad Request (400 Bad Request):**
```json
{
  "error": "missing app param"
}
```

**Server Error (500 Internal Server Error):**
```json
{
  "error": "internal server error"
}
```

## 🧪 Testing

### Run All Tests

```bash
go test ./...
```

### Run Specific Test Suites

```bash
# Unit tests (no database required)
go test ./internal/delivery -v

# Integration tests (requires database)
go test ./internal/campaigns -v

# All tests with coverage
go test ./... -cover
```

### Test Examples

The test suite includes comprehensive scenarios:

- ✅ Valid parameter validation
- ✅ Missing parameter handling
- ✅ Case-insensitive matching
- ✅ Complex targeting rules
- ✅ Multiple campaign matches
- ✅ No matches scenarios
- ✅ Performance testing
- ✅ Response format validation

## 📊 Sample Data

The application comes pre-loaded with sample campaigns and targeting rules:

### Campaigns
- **spotify**: Music streaming app
- **duolingo**: Language learning app  
- **subwaysurfer**: Mobile game

### Targeting Rules
- **spotify**: Available in US and Canada
- **duolingo**: Available on Android/iOS, excluded from US
- **subwaysurfer**: Available on Android for specific app

## 🔧 Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | Database host |
| `DB_PORT` | `5432` | Database port |
| `DB_NAME` | `targeting_db` | Database name |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `password` | Database password |
| `DB_SSL_MODE` | `disable` | SSL mode |

### Performance Considerations

- **Read-Heavy Optimization**: Database indexes on frequently queried columns
- **Connection Pooling**: Efficient database connection management
- **Query Optimization**: Single query with complex targeting logic
- **Caching Ready**: Architecture supports Redis/memcached integration

## 🚀 Deployment

### Docker Deployment

```bash
# Build and run
docker-compose up --build

# Production deployment
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### Kubernetes Deployment

```bash
# Apply Kubernetes manifests
kubectl apply -f k8s/
```

## 📈 Monitoring & Observability

### Health Checks
- Application health: `GET /healthz`
- Database connectivity monitoring
- Graceful shutdown handling

### Logging
- Structured logging with request IDs
- Performance metrics (request duration)
- Error tracking and debugging

### Metrics (Future Enhancement)
- Prometheus metrics integration
- Grafana dashboards
- Request rate monitoring
- Database performance metrics

## 📈 Monitoring (Prometheus & Grafana)

- Metrics endpoint: `GET /metrics`
- Exposed metrics:
  - `delivery_requests_total{status}`
  - `delivery_request_duration_seconds{status}`
  - `db_query_duration_seconds`

Start full stack with monitoring:

```bash
docker-compose up -d
# Prometheus: http://localhost:9090
# Grafana:    http://localhost:3000  (admin/admin)
```

## v2 API (go-kit)

- New endpoint (same behavior as v1) built with go-kit:

```http
GET /v2/delivery?app={app}&country={country}&os={os}
```

v1 routes remain for compatibility and tests.

## 🔍 Troubleshooting

### Common Issues

1. **Database Connection Failed**
   ```bash
   # Check if PostgreSQL is running
   docker-compose ps
   
   # Check logs
   docker-compose logs postgres
   ```

2. **No Campaigns Returned**
   ```bash
   # Verify data is seeded
   docker exec -it $(docker-compose ps -q postgres) psql -U postgres -d targeting_db -c "SELECT * FROM campaigns;"
   ```

3. **Port Already in Use**
   ```bash
   # Change port in docker-compose.yml
   ports:
     - "8081:8080"  # Use different host port
   ```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## 📝 License

This project is confidential and proprietary to GreedyGame Media Pvt Ltd.

---

**Built with ❤️ using Go, PostgreSQL, and Docker**
