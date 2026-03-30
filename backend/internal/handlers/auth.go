package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"tst-isp/internal/db"
	"tst-isp/internal/models"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password required")
		return
	}

	coll := db.DB().Collection("users")
	var user models.User
	if err := coll.FindOne(r.Context(), map[string]interface{}{"email": req.Email}).Decode(&user); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if user.Role != "super_admin" && user.Role != "sub_isp" {
		writeError(w, http.StatusForbidden, "role not allowed")
		return
	}
	if user.Role == "sub_isp" {
		if user.SubIspID == "" {
			writeError(w, http.StatusForbidden, "sub-isp not linked")
			return
		}
		var subIsp models.SubISP
		collSub := db.DB().Collection("sub_isps")
		if err := collSub.FindOne(r.Context(), map[string]interface{}{"_id": user.SubIspID}).Decode(&subIsp); err != nil {
			writeError(w, http.StatusForbidden, "sub-isp not found")
			return
		}
		if !subIsp.PaidUntil.IsZero() && subIsp.PaidUntil.Before(time.Now()) && subIsp.Status == "active" {
			subIsp.Status = "suspended"
			subIsp.UpdatedAt = time.Now()
			_, _ = collSub.UpdateOne(r.Context(), map[string]interface{}{"_id": user.SubIspID}, map[string]interface{}{
				"$set": map[string]interface{}{"status": subIsp.Status, "updated_at": subIsp.UpdatedAt},
			})
		}
		if subIsp.Status != "active" || subIsp.PaidUntil.IsZero() || subIsp.PaidUntil.Before(time.Now()) {
			writeError(w, http.StatusForbidden, "sub-isp not active")
			return
		}
	}

	token, err := signToken(user.Email, user.Role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(loginResponse{
		Token: token,
		Role:  user.Role,
	})
}

func signToken(email, role string) (string, error) {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		return "", errors.New("missing JWT_SECRET")
	}

	claims := jwt.MapClaims{
		"sub":  email,
		"role": role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}
