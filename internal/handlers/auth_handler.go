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

	return &AuthHandler{
		authService: authService,
		store:       sessions.NewCookieStore([]byte(sessionSecret)),
		templates:   templates,
	}, nil
}

// SignUp handles anonymous user signup
func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Render signup page
		http.ServeFile(w, r, "templates/signup.html")
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
	session, _ := h.store.Get(r, "feedback-session")
	session.Values["token"] = token
	session.Values["user_id"] = user.ID.String()
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
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
		http.ServeFile(w, r, "templates/login.html")
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
	session, _ := h.store.Get(r, "feedback-session")
	session.Values["token"] = req.Token
	session.Values["user_id"] = user.ID.String()
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
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

