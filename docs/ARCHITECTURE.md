# TST-ISP Core Architecture

## Overview
TST-ISP Core is a comprehensive ISP management system that handles user authentication, payment processing, RADIUS authentication, and COA (Change of Authorization) for MikroTik routers.

## Components

### Backend (Go)
- **Framework**: Standard library with custom routing
- **Database**: MongoDB for user/package data, MariaDB for RADIUS
- **Authentication**: JWT tokens with bcrypt password hashing
- **Payment**: M-Pesa STK Push integration
- **RADIUS**: FreeRADIUS server for user authentication
- **COA**: Disconnect users via MikroTik API

### Frontend (React + Vite)
- **UI**: Landing page, login, dashboards for owners and Sub-ISPs
- **State Management**: React hooks
- **API**: RESTful communication with backend

### Infrastructure
- **MongoDB**: User data, packages, payments
- **FreeRADIUS + MariaDB**: RADIUS authentication database
- **MikroTik Router**: Internet access control via RADIUS/COA

## Key Features
- Owner dashboard for system management
- Sub-ISP registration and management
- Payment processing with M-Pesa
- RADIUS-based authentication
- COA for session control
- Usage tracking and reporting

## Security
- JWT authentication
- Password hashing with bcrypt
- Role-based access control
- Input validation and sanitization
- HTTPS recommended for production

## Deployment
- Docker containers for services
- Environment-based configuration
- Local development with docker-compose

## MikroTik Integration
1. Configure router to use RADIUS server
2. Set COA address in environment
3. Users authenticated via RADIUS get internet access
4. Payments control access duration
5. COA used for disconnecting expired users

## API Endpoints
- `/api/auth/login`: User authentication
- `/api/mpesa/*`: Payment processing
- `/api/admin/*`: Owner management
- `/api/subisp/*`: Sub-ISP operations
- `/api/health`: Health check

## Environment Variables
See `.env.example` for required configuration.