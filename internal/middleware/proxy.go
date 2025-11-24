package middleware

import "net/http"

// IsHTTPS checks if the request is over HTTPS
// Handles Render and other proxy setups
func IsHTTPS(r *http.Request) bool {
	// Check X-Forwarded-Proto header (set by Render/proxies)
	if proto := r.Header.Get("X-Forwarded-Proto"); proto == "https" {
		return true
	}
	// Check X-Forwarded-Ssl header
	if ssl := r.Header.Get("X-Forwarded-Ssl"); ssl == "on" {
		return true
	}
	// Check direct TLS connection
	return r.TLS != nil
}

// SetSecureCookie sets Secure flag based on HTTPS detection
func SetSecureCookie(r *http.Request, options *http.Cookie) {
	if IsHTTPS(r) {
		options.Secure = true
	}
}

