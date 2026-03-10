# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

YogaLink API is a Go backend service for a yoga teacher-student matching platform. It provides RESTful APIs for user authentication, teacher profiles, availability management, booking requests, and real-name verification with face recognition.

## Commands

```bash
# Run the API server
go run ./cmd/api

# Build the API server
go build -v ./cmd/api

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Run a single test
go test -v -run TestName ./...

# Run linter (requires golangci-lint)
golangci-lint run --timeout=5m

# Run Docker build
docker build -f deployments/docker/Dockerfile -t yogalink-api:latest .
```

## Architecture

This project follows **clean architecture** with three layers:

```
cmd/
├── api/         # Main HTTP API server (Gin framework)
├── scheduler/   # Scheduled task service
└── worker/      # Background worker

internal/
├── handler/     # HTTP request handlers (Gin controllers)
├── service/     # Business logic layer
├── repository/  # Data access layer (GORM)
└── model/       # Database models
```

**Data Flow**: HTTP Request → Handler → Service → Repository → Database

## Key Components

- **Database**: MySQL with GORM ORM
- **Cache**: Redis for session/caching
- **Authentication**: JWT-based (Bearer token)
- **Config**: YAML-based via Viper (`configs/config.example.yaml`)

## API Structure

- `/v1/auth/*` - Authentication (register, login)
- `/v1/user/*` - User profile and verification
- `/v1/teacher/*` - Teacher profile, availability, locations, booking requests
- `/v1/student/*` - Student booking requests
- `/v1/teachers/nearby` - Public API for geo-based teacher search

## Important Models

- `User` - Base user with role (teacher/student)
- `TeacherProfile` - Teacher-specific profile (bio, specialties, hourly rate)
- `TeacherAvailability` - Time slots when teachers are available
- `TeachingLocation` - Teaching locations with geo-coordinates
- `BookingRequest` - Student booking requests with status workflow

## Database Migrations

Database migrations are handled automatically by GORM's AutoMigrate. Tables are created on server startup if they don't exist.
