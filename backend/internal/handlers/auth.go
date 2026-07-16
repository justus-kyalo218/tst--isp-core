package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"

	"tst-isp/internal/db"
	"tst-isp/internal/models"
	"tst-isp/pkg/logger"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

type adminRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type forgotPasswordResponse struct {
	Message string `json:"message"`
}

type resetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type passwordReset struct {
	Email     string    `bson:"email"`
	TokenHash string    `bson:"token_hash"`
	ExpiresAt time.Time `bson:"expires_at"`
	CreatedAt time.Time `bson:"created_at"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode login request: %v", err)
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	coll := db.DB().Collection("users")
	var user models.User
	if err := coll.FindOne(r.Context(), map[string]interface{}{"email": req.Email}).Decode(&user); err != nil {
		logger.Warn("login attempt for non-existent user: %s", req.Email)
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		logger.Warn("invalid password for user: %s", req.Email)
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if user.Role != "super_admin" && user.Role != "sub_isp" {
		logger.Warn("login attempt with invalid role: %s for user %s", user.Role, req.Email)
		writeError(w, http.StatusForbidden, "role not allowed")
		return
	}
	if user.Role == "sub_isp" {
		if user.SubIspID == "" {
			logger.Warn("sub-isp user not linked: %s", req.Email)
			writeError(w, http.StatusForbidden, "sub-isp not linked")
			return
		}
		var subIsp models.SubISP
		collSub := db.DB().Collection("sub_isps")
		if err := collSub.FindOne(r.Context(), map[string]interface{}{"_id": user.SubIspID}).Decode(&subIsp); err != nil {
			logger.Warn("sub-isp not found for user: %s", req.Email)
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
			logger.Warn("inactive sub-isp login attempt: %s", req.Email)
			writeError(w, http.StatusForbidden, "sub-isp not active")
			return
		}
	}

	token, err := signToken(user.Email, user.Role)
	if err != nil {
		logger.Error("failed to generate token for user %s: %v", user.Email, err)
		writeError(w, http.StatusInternalServerError, "token error")
		return
	}

	logger.Info("successful login for user: %s", user.Email)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(loginResponse{
		Token: token,
		Role:  user.Role,
	})
}

func AdminRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req adminRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode admin registration request: %v", err)
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		writeError(w, http.StatusBadRequest, "email is invalid")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	coll := db.DB().Collection("users")
	adminCount, err := coll.CountDocuments(r.Context(), bson.M{"role": "super_admin"})
	if err != nil {
		logger.Error("failed to count admins: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to register admin")
		return
	}
	if adminCount > 0 {
		writeError(w, http.StatusForbidden, "first admin already exists")
		return
	}

	if err := createSuperAdmin(r.Context(), req.Email, req.Password); err != nil {
		if errors.Is(err, errEmailRegistered) {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}
		logger.Error("failed to create first admin: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to register admin")
		return
	}

	token, err := signToken(req.Email, "super_admin")
	if err != nil {
		logger.Error("failed to generate token for admin %s: %v", req.Email, err)
		writeError(w, http.StatusInternalServerError, "token error")
		return
	}

	logger.Info("registered admin: %s", req.Email)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(loginResponse{
		Token: token,
		Role:  "super_admin",
	})
}

func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		writeError(w, http.StatusBadRequest, "email is invalid")
		return
	}

	coll := db.DB().Collection("users")
	var user models.User
	if err := coll.FindOne(r.Context(), bson.M{"email": req.Email}).Decode(&user); err == nil {
		if user.Role == "super_admin" || user.Role == "sub_isp" {
			token, err := createPasswordReset(r.Context(), user.Email)
			if err != nil {
				logger.Error("failed to create password reset token for %s: %v", user.Email, err)
			} else {
				resetURL := passwordResetURL(r, token)
				if err := sendPasswordResetEmail(user.Email, resetURL); err != nil {
					logger.Error("failed to send password reset email to %s: %v", user.Email, err)
				}
				logger.Warn("password reset requested for %s account: %s", user.Role, req.Email)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(forgotPasswordResponse{
		Message: "If this email matches an admin or Sub-ISP account, a password reset link has been sent.",
	})
}

func ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "token and password are required")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	resets := db.DB().Collection("password_resets")
	var reset passwordReset
	err := resets.FindOne(r.Context(), bson.M{
		"token_hash": hashResetToken(req.Token),
		"expires_at": bson.M{
			"$gt": time.Now(),
		},
	}).Decode(&reset)
	if err != nil {
		writeError(w, http.StatusBadRequest, "reset link is invalid or expired")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("failed to hash reset password for %s: %v", reset.Email, err)
		writeError(w, http.StatusInternalServerError, "failed to reset password")
		return
	}

	result, err := db.DB().Collection("users").UpdateOne(r.Context(), bson.M{"email": reset.Email}, bson.M{
		"$set": bson.M{
			"password":   string(hash),
			"updated_at": time.Now(),
		},
	})
	if err != nil {
		logger.Error("failed to reset password for %s: %v", reset.Email, err)
		writeError(w, http.StatusInternalServerError, "failed to reset password")
		return
	}
	if result.MatchedCount == 0 {
		writeError(w, http.StatusBadRequest, "reset link is invalid or expired")
		return
	}

	_, _ = resets.DeleteMany(r.Context(), bson.M{"email": reset.Email})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Password reset successfully. You can now log in with your new password.",
	})
}

var errEmailRegistered = errors.New("email already registered")

func createSuperAdmin(ctx context.Context, email, password string) error {
	coll := db.DB().Collection("users")
	if count, err := coll.CountDocuments(ctx, bson.M{"email": email}); err != nil {
		return err
	} else if count > 0 {
		return errEmailRegistered
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = coll.InsertOne(ctx, models.User{
		Email:     email,
		Password:  string(hash),
		Role:      "super_admin",
		CreatedAt: now,
		UpdatedAt: now,
	})
	return err
}

func createPasswordReset(ctx context.Context, email string) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	now := time.Now()
	coll := db.DB().Collection("password_resets")
	_, _ = coll.DeleteMany(ctx, bson.M{"email": email})
	_, err := coll.InsertOne(ctx, passwordReset{
		Email:     email,
		TokenHash: hashResetToken(token),
		ExpiresAt: now.Add(time.Hour),
		CreatedAt: now,
	})
	return token, err
}

func hashResetToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func passwordResetURL(r *http.Request, token string) string {
	appURL := strings.TrimRight(strings.TrimSpace(os.Getenv("APP_URL")), "/")
	if appURL == "" {
		scheme := "http"
		if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
			scheme = "https"
		}
		host := r.Host
		if forwardedHost := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); forwardedHost != "" {
			host = forwardedHost
		}
		appURL = scheme + "://" + host
	}
	return appURL + "/reset-password?reset_token=" + token
}

func sendPasswordResetEmail(to, resetURL string) error {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	user := strings.TrimSpace(os.Getenv("SMTP_USER"))
	pass := os.Getenv("SMTP_PASS")
	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))

	if host == "" || port == "" || from == "" {
		logger.Warn("SMTP is not configured; password reset link for %s: %s", to, resetURL)
		return nil
	}
	if _, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	auth := smtp.Auth(nil)
	if user != "" || pass != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}

	subject := "Reset your TST ISP password"
	body := "Use this link to reset your password. It expires in 1 hour:\r\n\r\n" + resetURL + "\r\n\r\nIf you did not request this, you can ignore this email.\r\n"
	msg := strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	return smtp.SendMail(host+":"+port, auth, from, []string{to}, []byte(msg))
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
