# Directory Guide: Backend vs Frontend

## 📁 Visual Directory Structure

```
feedback-sys/
│
├── 🔴 BACKEND (Go Server Code)
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              ← Server entry point
│   │
│   └── internal/                    ← All backend logic
│       ├── config/                  ← Configuration
│       ├── database/                 ← DB connection
│       ├── handlers/                 ← HTTP handlers (serve HTML)
│       ├── llm/                      ← LLM integration
│       ├── middleware/               ← Auth, tracing
│       ├── models/                   ← Data models
│       ├── repository/               ← Database queries
│       └── service/                  ← Business logic
│
├── 🟢 FRONTEND (User Interface)
│   ├── templates/                    ← HTML pages
│   │   ├── dashboard.html
│   │   ├── mood_dashboard.html
│   │   ├── quiz_list.html
│   │   └── ...
│   │
│   └── static/                       ← CSS, images
│       └── css/
│           └── style.css
│
├── 🔵 DATABASE
│   └── migrations/                   ← Database schema
│       ├── 001_initial_schema.up.sql
│       ├── 002_mood_tracking.up.sql
│       └── 003_add_quiz_questions.up.sql
│
└── ⚙️ CONFIG (Deployment)
    ├── render.yaml                   ← Render config
    ├── Dockerfile                    ← Docker config
    ├── .env                          ← Environment vars
    └── go.mod                        ← Go dependencies
```

## 🎯 Quick Reference

### Backend Files (`.go` extension)
- **Location**: `cmd/`, `internal/`
- **What they do**: 
  - Handle HTTP requests
  - Process business logic
  - Query database
  - Generate HTML responses

### Frontend Files (`.html`, `.css`)
- **Location**: `templates/`, `static/`
- **What they do**:
  - Define page layout
  - Style pages
  - User interactions

### Database Files (`.sql`)
- **Location**: `migrations/`
- **What they do**: Create database tables

## 🔄 Request Flow

```
1. User visits /mood
   ↓
2. Backend (handlers/mood_handler.go) receives request
   ↓
3. Backend queries database (repository/mood_repository.go)
   ↓
4. Backend processes data (service/mood_service.go)
   ↓
5. Backend renders template (templates/mood_dashboard.html)
   ↓
6. Backend sends HTML + CSS to browser
   ↓
7. User sees the page
```

## 🚀 Deployment

**For Render, you deploy EVERYTHING:**
- ✅ All backend code
- ✅ All frontend templates
- ✅ All static files
- ✅ All migrations

**Render will:**
1. Build Go backend
2. Copy templates and static files
3. Run the server
4. Server serves everything

## 💡 Key Insight

**This is NOT a separate frontend/backend split!**

It's a **monolithic server-side rendered** application:
- Go backend generates HTML
- Templates are embedded in the backend
- Everything runs as one process
- Everything deploys together
