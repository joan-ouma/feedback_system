package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/joan/feedback-sys/internal/models"
)

type contextKey string

const userContextKey contextKey = "user"

// AuthMiddleware handles authentication using session tokens
type AuthMiddleware struct {
	store *sessions.CookieStore
}

func NewAuthMiddleware(sessionSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		store: sessions.NewCookieStore([]byte(sessionSecret)),
	}
}

// RequireAuth middleware ensures user is authenticated
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := m.store.Get(r, "feedback-session")
		token, ok := session.Values["token"].(string)

		if !ok || token == "" {
			// Redirect to signup/login page
			http.Redirect(w, r, "/signup", http.StatusSeeOther)
			return
		}

		// Store token in context for handlers to use
		ctx := context.WithValue(r.Context(), "token", token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth middleware adds user to context if authenticated, but doesn't require it
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := m.store.Get(r, "feedback-session")
		if token, ok := session.Values["token"].(string); ok && token != "" {
			ctx := context.WithValue(r.Context(), "token", token)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext retrieves the user from request context
func GetUserFromContext(ctx context.Context) *models.User {
	if user, ok := ctx.Value(userContextKey).(*models.User); ok {
		return user
	}
	return nil
}

// GetTokenFromContext retrieves the token from request context
func GetTokenFromContext(ctx context.Context) string {
	if token, ok := ctx.Value("token").(string); ok {
		return token
	}
	return ""
}

