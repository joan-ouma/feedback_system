package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joan/feedback-sys/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	store       *sessions.CookieStore
	templates   *template.Template
}

type TokenDisplayData struct {
	Token string
}

func NewAuthHandler(authService *service.AuthService, sessionSecret string, templateDir string) (*AuthHandler, error) {
	tmpl := template.New("")
	
	// Load template files
	pattern := filepath.Join(templateDir, "*.html")
	templates, err := tmpl.ParseGlob(pattern)
	if err != nil {
		return nil, err
	}

	store := sessions.NewCookieStore([]byte(sessionSecret))
	
	// Configure cookie options for production (HTTPS)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30, // 30 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode, // Works with HTTPS and allows cross-site navigation
	}

	return &AuthHandler{
		authService: authService,
		store:       store,
		templates:   templates,
	}, nil
}

// SignUp handles anonymous user signup
func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Render signup page using template
		w.Header().Set("Content-Type", "text/html")
		if err := h.templates.ExecuteTemplate(w, "signup.html", nil); err != nil {
			// Fallback to file serve if template not found
			http.ServeFile(w, r, "templates/signup.html")
			return
		}
		return
	}

	// Handle POST request - support both form and JSON
	var req struct {
		DisplayName string `json:"display_name"`
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
	} else {
		// Handle form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		req.DisplayName = r.FormValue("display_name")
	}

	user, token, err := h.authService.SignUp(r.Context(), req.DisplayName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store token in session
	session, err := h.store.Get(r, "feedback-session")
	if err != nil {
		// If session can't be retrieved, create a new one
		session, _ = h.store.New(r, "feedback-session")
	}
	session.Values["token"] = token
	session.Values["user_id"] = user.ID.String()
	
	// Set cookie options for production (HTTPS detection)
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" || r.Header.Get("X-Forwarded-Ssl") == "on" {
		session.Options.Secure = true
		session.Options.SameSite = http.SameSiteLaxMode
	}
	
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return HTMX response - show token display instead of redirecting
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		tokenData := TokenDisplayData{Token: token}
		if err := h.templates.ExecuteTemplate(w, "token_display.html", tokenData); err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	
	// Fallback: redirect for non-HTMX requests
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Login handles user login with token
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Render login page using template
		w.Header().Set("Content-Type", "text/html")
		if err := h.templates.ExecuteTemplate(w, "login.html", nil); err != nil {
			// Fallback to file serve if template not found
			http.ServeFile(w, r, "templates/login.html")
			return
		}
		return
	}

	var req struct {
		Token string `json:"token"`
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
	} else {
		// Handle form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		req.Token = r.FormValue("token")
	}

	user, err := h.authService.Authenticate(r.Context(), req.Token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Store token in session
	session, err := h.store.Get(r, "feedback-session")
	if err != nil {
		// If session can't be retrieved, create a new one
		session, _ = h.store.New(r, "feedback-session")
	}
	session.Values["token"] = req.Token
	session.Values["user_id"] = user.ID.String()
	
	// Set cookie options for production (HTTPS detection)
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" || r.Header.Get("X-Forwarded-Ssl") == "on" {
		session.Options.Secure = true
		session.Options.SameSite = http.SameSiteLaxMode
	}
	
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "feedback-session")
	session.Values["token"] = ""
	session.Values["user_id"] = ""
	session.Options.MaxAge = -1
	session.Save(r, w)

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/signup")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/signup", http.StatusSeeOther)
}

func (h *AuthHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/signup", h.SignUp).Methods("GET", "POST")
	router.HandleFunc("/login", h.Login).Methods("GET", "POST")
	router.HandleFunc("/logout", h.Logout).Methods("POST")
}

