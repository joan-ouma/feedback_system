# Application Architecture: Backend vs Frontend

This application uses a **server-side rendered (SSR)** architecture with HTMX, which means the backend serves HTML directly. Here's how it's organized:

## 🎯 Quick Answer

**Everything is deployed together** - This is a monolithic application where:
- **Backend (Go)** serves the frontend (HTML templates)
- **Frontend (HTML/CSS/JS)** is embedded in the backend
- No separate frontend build step needed

## 📁 Directory Structure Explained

### 🔴 Backend Files (Go Code)

These are the **server-side** files that handle logic, database, and API:

```
cmd/server/              # Backend entry point (main.go)
internal/
  ├── config/           # Backend: Configuration management
  ├── database/         # Backend: Database connection
  ├── handlers/         # Backend: HTTP request handlers (serve HTML)
  ├── llm/              # Backend: LLM API integration
  ├── middleware/       # Backend: Auth, tracing middleware
  ├── models/           # Backend: Data models
  ├── repository/       # Backend: Database queries
  └── service/          # Backend: Business logic
migrations/             # Backend: Database schema
go.mod                  # Backend: Go dependencies
go.sum                  # Backend: Go dependencies lock
```

**What they do:**
- Handle HTTP requests
- Process business logic
- Interact with database
- Generate HTML responses
- Manage authentication

### 🟢 Frontend Files (User Interface)

These are the **client-side** files that users see and interact with:

```
templates/              # Frontend: HTML templates (served by backend)
  ├── dashboard.html
  ├── mood_dashboard.html
  ├── quiz_list.html
  ├── consultation.html
  └── ...

static/                 # Frontend: CSS, images, client-side JS
  └── css/
      └── style.css
```

**What they do:**
- Define the visual layout
- Style the pages (CSS)
- Handle user interactions (HTMX)
- Display data to users

### 🔵 Hybrid Files (Both)

These files are used by both or configure the deployment:

```
render.yaml             # Deployment config (both)
Dockerfile              # Deployment config (both)
.env                    # Configuration (backend reads it)
```

## 🏗️ How It Works Together

```
User Browser
    ↓
    ↓ HTTP Request
    ↓
Go Backend (cmd/server/main.go)
    ↓
    ├─→ Handlers (internal/handlers/)
    │   ├─→ Process request
    │   ├─→ Call Services (internal/service/)
    │   ├─→ Query Database (internal/repository/)
    │   └─→ Render HTML Template (templates/)
    │
    └─→ Return HTML + CSS + HTMX
        ↓
User Browser (displays page)
```

## 📊 Request Flow Example

**When user visits `/mood`:**

1. **Browser** → Sends GET request to `/mood`
2. **Backend Handler** (`internal/handlers/mood_handler.go`)
   - Authenticates user
   - Calls `MoodService`
   - Gets data from database
3. **Backend** → Renders `templates/mood_dashboard.html`
   - Fills template with data
   - Includes CSS from `static/css/style.css`
   - Includes HTMX script
4. **Browser** → Receives complete HTML page
5. **Frontend** → HTMX handles interactions (no page reload)

## 🎨 Frontend Technologies

- **HTML Templates**: Go templates in `templates/`
- **CSS**: Stylesheets in `static/css/`
- **HTMX**: Loaded from CDN (no local files needed)
- **JavaScript**: Minimal JS in templates for interactions

## ⚙️ Backend Technologies

- **Go**: All `.go` files
- **PostgreSQL**: Database (via `internal/database/`)
- **HTMX**: Server-side rendering (backend generates HTML)
- **Gorilla Mux**: HTTP router

## 🚀 Deployment

**For Render deployment, you deploy EVERYTHING together:**

```
Your Repository
├── Backend (Go code) ✅
├── Frontend (Templates + CSS) ✅
└── Config files ✅
```

Render will:
1. Build the Go backend (`go build`)
2. Copy templates and static files
3. Run the server
4. Server serves both API and HTML pages

## 🔍 How to Identify

### Backend Files:
- ✅ File extension: `.go`
- ✅ Location: `cmd/`, `internal/`
- ✅ Contains: Business logic, database code, API handlers

### Frontend Files:
- ✅ File extension: `.html`, `.css`, `.js`
- ✅ Location: `templates/`, `static/`
- ✅ Contains: UI, styling, user interactions

### Both:
- ✅ Configuration files (`.yaml`, `Dockerfile`)
- ✅ Templates are rendered by backend but contain frontend code

## 💡 Key Point

**This is NOT a separate frontend/backend architecture.**

Instead, it's:
- **Monolithic**: One application serves everything
- **Server-Side Rendered**: Backend generates HTML
- **HTMX Enhanced**: Frontend uses HTMX for dynamic updates without full page reloads

## 📝 Summary

| Component | Location | Purpose |
|-----------|---------|---------|
| **Backend Logic** | `internal/`, `cmd/` | Handles requests, processes data |
| **Frontend UI** | `templates/`, `static/` | What users see and interact with |
| **Database** | `migrations/` | Schema definitions |
| **Config** | `.env`, `render.yaml` | Environment settings |

**Everything gets deployed together as one application!**

