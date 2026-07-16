package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"

	"tst-isp/pkg/logger"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {

		return ""
	}
	var msgs []string
	for _, e := range ve {
		msgs = append(msgs, e.Field+": "+e.Message)
	}
	return strings.Join(msgs, "; ")
}

func ValidateLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn("invalid JSON in login request")
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		var errors ValidationErrors

		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		if req.Email == "" {
			errors = append(errors, ValidationError{Field: "email", Message: "required"})
		} else if !isValidEmail(req.Email) {
			errors = append(errors, ValidationError{Field: "email", Message: "invalid format"})
		}

		if req.Password == "" {
			errors = append(errors, ValidationError{Field: "password", Message: "required"})
		} else if len(req.Password) < 6 {
			errors = append(errors, ValidationError{Field: "password", Message: "must be at least 6 characters"})
		}

		if len(errors) > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"errors": errors})
			return
		}

		// Re-encode body for next handler
		data, _ := json.Marshal(req)
		r.Body = io.NopCloser(strings.NewReader(string(data)))

		next.ServeHTTP(w, r)
	})
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
