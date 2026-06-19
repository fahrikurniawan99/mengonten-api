# Mengonten API - Setup Guide

## Prerequisites
- Go 1.20+ (ideally 1.25+)
- PostgreSQL running
- Git

## Initial Setup

### 1. Fix Go Environment (if needed)
Pastikan GOROOT mengarah ke instalasi Go yang benar:
```bash
go env GOROOT
```

### 2. Install Dependencies
```bash
$env:GOPROXY="https://proxy.golang.org,direct"
$env:GOSUMDB="off"
$env:GO111MODULE="on"
go mod tidy
```

### 3. Setup Environment Variables
Copy `.env.example` ke `.env` dan sesuaikan:
```bash
cp .env.example .env
```

Edit `.env` dengan PostgreSQL credentials:
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=mengonten_db
JWT_SECRET=your-secret-key-change-in-production
GIN_MODE=debug
PORT=8080
```

### 4. Create Database
```sql
CREATE DATABASE mengonten_db;
```

### 5. Generate Swagger Documentation
Install swag CLI:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

Generate docs:
```bash
swag init
```

This creates `docs/` folder dengan `docs.go`, `swagger.json`, dan `swagger.yaml`.

### 6. Run Application
```bash
go run main.go
```

Server akan berjalan di `http://localhost:8080`

## API Documentation

### Swagger UI
Akses dokumentasi API di: `http://localhost:8080/swagger/index.html`

### Available Endpoints

#### Authentication Routes
- `POST /api/auth/register` - Register user baru
- `POST /api/auth/login` - Login dan dapatkan JWT token
- `POST /api/auth/logout` - Logout user

#### Protected Routes (memerlukan JWT Bearer token)
- `GET /api/profile` - Dapatkan profile user

### JWT Authentication
Untuk akses protected routes, tambahkan header:
```
Authorization: Bearer <your_jwt_token>
```

## Project Structure
```
├── main.go              # Entry point
├── config/
│   ├── database.go      # Database connection
│   └── jwt.go           # JWT configuration
├── models/
│   └── user.go          # User model
├── routes/
│   ├── auth.go          # Authentication handlers
│   └── routes.go        # Route registration
├── middleware/
│   └── auth.go          # JWT middleware
├── utils/
│   └── response.go      # Response utilities
├── docs.go              # Swagger documentation
├── go.mod               # Go modules
├── .env.example         # Environment template
└── .gitignore           # Git ignore file
```

## Development

### Adding New Routes
1. Create handler di `routes/`
2. Add swagger comments ke handler
3. Register route di `routes/routes.go`
4. Run `swag init` untuk update docs

### Swagger Annotations Format
```go
// @Summary Brief description
// @Description Longer description
// @Tags TagName
// @Accept json
// @Produce json
// @Param param-name query/path/body type required "description"
// @Success 200 {object} ResponseType "message"
// @Failure 400 {object} utils.Response "error"
// @Router /api/route [http-method]
func HandlerName(c *gin.Context) {
    // implementation
}
```

## Troubleshooting

### Issue: "Failed to connect to database"
- Pastikan PostgreSQL running
- Check `.env` credentials
- Database sudah di-create

### Issue: "GOROOT pointing to test directory"
Set GOROOT ke Go installation yang benar:
```bash
$env:GOROOT="C:\Program Files\Go"
```

### Issue: Swagger docs not generated
Run `swag init` dan ensure swagger comments format benar.

## Build for Production
```bash
go build -o mengonten-api.exe
```

## Testing
Testing setup coming soon...
