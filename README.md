# Mengonten API

API untuk auto-clipping video YouTube dengan AI genre detection dan content analysis.

## Features

- **User Authentication** - JWT-based authentication system
- **YouTube Auto-Clipping** - Automatic video clipping with AI analysis
- **Genre Detection** - AI-powered video genre classification (Comedy, Educational, Sports, Gaming, Music, etc.)
- **Smart Segment Finding** - Intelligent selection of interesting segments based on content
- **Cloud Upload** - Automatic upload to Cloudinary

## Tech Stack

- **Language:** Go 1.25.0+
- **Framework:** Gin
- **Database:** PostgreSQL + GORM
- **AI Services:** OpenAI Whisper (transcription) + GPT-4 (analysis)
- **Video Processing:** FFmpeg + yt-dlp
- **Cloud Storage:** Cloudinary
- **Documentation:** Swagger/OpenAPI

## Prerequisites

- Go 1.23.1+
- PostgreSQL 13+
- Python 3.10+ (for yt-dlp)
- FFmpeg
- yt-dlp

## Installation

### 1. Clone Repository

```bash
git clone https://github.com/fahrikurniawan99/mengonten-api.git
cd mengonten-api
```

### 2. Install Dependencies

```powershell
$env:GO111MODULE="on"
go mod tidy
```

### 3. Install External Tools

**FFmpeg:**
```powershell
choco install ffmpeg
```

**yt-dlp:**
```powershell
pip install yt-dlp
```

### 4. Setup Environment Variables

Copy `.env.example` to `.env`:

```powershell
Copy-Item .env.example .env
```

Edit `.env` with your credentials:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=mengonten_db

JWT_SECRET=your-jwt-secret

OPENAI_API_KEY=your-openai-api-key

CLOUDINARY_NAME=your-cloudinary-name
CLOUDINARY_KEY=your-cloudinary-api-key
CLOUDINARY_SECRET=your-cloudinary-api-secret

GIN_MODE=debug
PORT=8080
```

### 5. Create Database

```sql
CREATE DATABASE mengonten_db;
```

### 6. Run Application

```powershell
$env:GOROOT="C:\Program Files\Go"
$env:GO111MODULE="on"
go run main.go
```

Server akan berjalan di `http://localhost:8080`

## API Documentation

Access Swagger UI at: `http://localhost:8080/swagger/index.html`

## API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register new user |
| POST | `/api/auth/login` | Login and get JWT token |
| POST | `/api/auth/logout` | Logout user |
| GET | `/api/profile` | Get user profile (protected) |

### YouTube Auto-Clipping

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/youtube/submit` | Submit YouTube video for processing |
| GET | `/api/youtube/{video_id}` | Get video details and clips |
| GET | `/api/youtube/jobs/{job_id}` | Check processing status |

## Usage Examples

### Register User

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@test.com","username":"testuser","password":"password123"}'
```

### Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@test.com","password":"password123"}'
```

### Submit YouTube Video

```bash
curl -X POST http://localhost:8080/api/youtube/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"youtube_url":"https://www.youtube.com/watch?v=VIDEO_ID"}'
```

### Check Processing Status

```bash
curl -X GET http://localhost:8080/api/youtube/jobs/JOB_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Processing Pipeline

1. **Download** - Video downloaded from YouTube using yt-dlp
2. **Transcribe** - Audio extracted and transcribed using OpenAI Whisper
3. **Analyze Genre** - Content analyzed using GPT-4 to determine genre
4. **Find Segments** - Interesting segments identified based on genre
5. **Cut Clips** - Video segments cut using FFmpeg
6. **Upload** - Clips uploaded to Cloudinary

## Project Structure

```
mengonten-api/
├── config/           # Configuration files
│   ├── database.go   # Database connection
│   ├── jwt.go        # JWT configuration
│   └── external_apis.go  # External API config
├── middleware/       # HTTP middleware
│   └── auth.go       # JWT authentication middleware
├── models/           # Database models
│   ├── user.go       # User model
│   └── youtube_video.go  # YouTube models
├── routes/           # API routes
│   ├── auth.go       # Authentication routes
│   ├── youtube.go    # YouTube routes
│   └── routes.go     # Route registration
├── utils/            # Utility functions
│   └── response.go   # API response utilities
├── worker/           # Background workers
│   ├── youtube_processor.go    # Main orchestrator
│   ├── youtube_downloader.go   # Video download
│   ├── transcript_generator.go # Transcription
│   ├── genre_analyzer.go       # Genre detection
│   ├── video_clipper.go        # Video cutting
│   └── cloudinary_uploader.go  # Cloud upload
├── docs/             # Swagger documentation
├── main.go           # Application entry point
├── go.mod            # Go modules
├── .env.example      # Environment template
└── README.md         # This file
```

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DB_HOST` | PostgreSQL host | Yes |
| `DB_PORT` | PostgreSQL port | Yes |
| `DB_USER` | PostgreSQL user | Yes |
| `DB_PASSWORD` | PostgreSQL password | Yes |
| `DB_NAME` | Database name | Yes |
| `JWT_SECRET` | JWT signing secret | Yes |
| `OPENAI_API_KEY` | OpenAI API key | Yes |
| `CLOUDINARY_NAME` | Cloudinary cloud name | Yes |
| `CLOUDINARY_KEY` | Cloudinary API key | Yes |
| `CLOUDINARY_SECRET` | Cloudinary API secret | Yes |
| `GIN_MODE` | Gin mode (debug/release) | No |
| `PORT` | Server port (default: 8080) | No |

## Troubleshooting

### yt-dlp HTTP 403 Error

1. Update Python to 3.10+
2. Update yt-dlp: `pip install --upgrade yt-dlp`
3. Try with a different public YouTube video
4. Some videos may be geo-blocked or age-restricted

### Database Connection Error

1. Verify PostgreSQL is running
2. Check credentials in `.env`
3. Ensure database exists

### Go Compilation Errors

1. Verify Go version: `go version` (should be 1.23.1+)
2. Run `go mod tidy`
3. Set GOROOT if needed: `$env:GOROOT="C:\Program Files\Go"`

## License

MIT License

## Author

Fahri Kurniawan
