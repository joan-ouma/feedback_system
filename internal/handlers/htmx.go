package handlers

import (
	"html/template"
	"io"
	"path/filepath"
)

// HTMX helper functions
type HTMX struct {
	templates *template.Template
}

func NewHTMX(templateDir string) (*HTMX, error) {
	tmpl := template.New("").Funcs(template.FuncMap{
		"formatDate": func(t interface{}) string {
			// This will be handled in templates
			return ""
		},
	})

	// Load all templates
	pattern := filepath.Join(templateDir, "*.html")
	_, err := tmpl.ParseGlob(pattern)
	if err != nil {
		return nil, err
	}

	return &HTMX{templates: tmpl}, nil
}

// RenderTemplate renders a template with the given data
func (h *HTMX) RenderTemplate(w io.Writer, name string, data interface{}) error {
	return h.templates.ExecuteTemplate(w, name, data)
}

// IsHTMXRequest checks if the request is from HTMX
func IsHTMXRequest(r interface{}) bool {
	if req, ok := r.(interface{ Header() map[string][]string }); ok {
		headers := req.Header()
		if hxRequest, exists := headers["Hx-Request"]; exists {
			return len(hxRequest) > 0 && hxRequest[0] == "true"
		}
		// Also check HX-Request (case variations)
		if hxRequest, exists := headers["HX-Request"]; exists {
			return len(hxRequest) > 0 && hxRequest[0] == "true"
		}
	}
	return false
}

// SetHTMXHeaders sets HTMX response headers
func SetHTMXHeaders(w interface{}, redirect string) {
	if resp, ok := w.(interface{ Header() map[string][]string }); ok {
		headers := resp.Header()
		if redirect != "" {
			headers["HX-Redirect"] = []string{redirect}
		}
		headers["Content-Type"] = []string{"text/html"}
	}
}

