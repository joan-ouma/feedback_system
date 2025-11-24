<<<<<<< HEAD
# feedback_system
An feedback system to help students access guidance and counselling anonymously
=======
# Campus Mental Health Support System

A production-ready web application built with Go, HTMX, PostgreSQL, and OpenTelemetry that helps students manage their mental health by anonymously submitting feedback and consulting with an AI counselor.

## Features

- **Anonymous Authentication**: Clever signup system using cryptographic tokens (no email/phone required)
- **Feedback Submission**: Students can submit anonymous feedback about campus issues, mental health concerns, etc.
- **AI Consultation**: Integrated LLM with complex system prompts for mental health counseling
- **OpenTelemetry**: Full observability with distributed tracing
- **Clean Architecture**: Standard Go project structure with separation of concerns

## Architecture

The application follows a clean, layered architecture:

```
cmd/server/          # Application entry point
internal/
  ├── config/        # Configuration management
  ├── database/      # Database connection and utilities
  ├── handlers/      # HTTP handlers (presentation layer)
  ├── llm/           # LLM client integration
  ├── middleware/    # HTTP middleware (auth, tracing)
  ├── models/        # Domain models
  ├── repository/    # Data access layer
  └── service/       # Business logic layer
migrations/          # Database migrations
static/              # Static assets (CSS, JS)
templates/           # HTML templates with HTMX
```

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Jaeger (for OpenTelemetry tracing, optional)
- LLM API access (OpenAI-compatible API)

## Setup

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd feedback-sys
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Set up PostgreSQL database**:
   ```bash
   createdb feedback_sys
   ```

4. **Run migrations**:
   ```bash
   psql feedback_sys < migrations/001_initial_schema.up.sql
   ```

5. **Configure environment variables**:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

   Required environment variables:
   - `DATABASE_URL`: PostgreSQL connection string
   - `SESSION_SECRET`: Secret key for session encryption
   - `LLM_API_KEY`: API key for LLM service
   - `LLM_API_URL`: LLM API endpoint (default: https://api.openai.com/v1)
   - `LLM_MODEL`: Model name (default: gpt-4)
   - `JAEGER_ENDPOINT`: Jaeger collector endpoint (optional)
   - `ENABLE_TRACING`: Enable/disable tracing (default: true)

6. **Run the application**:
   ```bash
   go run cmd/server/main.go
   ```

   Or build and run:
   ```bash
   go build -o bin/server cmd/server/main.go
   ./bin/server
   ```

7. **Access the application**:
   Open your browser and navigate to `http://localhost:8080`

## Anonymous Signup System

The application uses a clever anonymous signup system:

1. Users sign up without providing email or phone
2. A cryptographically secure token is generated (32 random bytes, base64 encoded)
3. The token is stored in the user's session
4. Users can save their token to access their account from other devices
5. All user data is completely anonymous

## API Endpoints

### Authentication
- `GET /signup` - Signup page
- `POST /signup` - Create anonymous account
- `GET /login` - Login page
- `POST /login` - Login with token
- `POST /logout` - Logout

### Feedback
- `GET /dashboard` - Dashboard page
- `POST /api/feedback` - Submit feedback
- `GET /api/feedback` - Get user's feedbacks
- `GET /api/feedback/{id}` - Get specific feedback

### Consultation
- `GET /consultation` - Consultation chat page
- `POST /api/consultation/session` - Start new session
- `POST /api/consultation/message` - Send message to LLM
- `GET /api/consultation/session/{session_id}/history` - Get session history

## LLM Integration

The application integrates with OpenAI-compatible APIs. The LLM service includes:

- Complex system prompts designed for mental health counseling
- Conversation history management
- Safety considerations and boundaries
- Campus resource suggestions

## OpenTelemetry

The application is instrumented with OpenTelemetry for distributed tracing:

- HTTP request tracing
- Database query tracing
- Service-level tracing
- LLM API call tracing

To view traces, ensure Jaeger is running and configured in your `.env` file.

## Development

### Running Tests
```bash
go test ./...
```

### Database Migrations
- Up: `psql feedback_sys < migrations/001_initial_schema.up.sql`
- Down: `psql feedback_sys < migrations/001_initial_schema.down.sql`

### Building for Production
```bash
go build -ldflags="-s -w" -o bin/server cmd/server/main.go
```

## Production Considerations

1. **Security**:
   - Use strong `SESSION_SECRET` in production
   - Enable HTTPS/TLS
   - Configure CORS appropriately
   - Rate limiting for API endpoints

2. **Database**:
   - Use connection pooling (already configured)
   - Regular backups
   - Index optimization

3. **Monitoring**:
   - OpenTelemetry traces
   - Application logs
   - Database performance metrics

4. **Scaling**:
   - Stateless application design (sessions in cookies)
   - Horizontal scaling ready
   - Database read replicas for read-heavy workloads

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

>>>>>>> 61da7d1 (Add source files for my feedback system)
