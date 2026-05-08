package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"tst-isp/internal/db"
	"tst-isp/internal/models"
	"tst-isp/internal/routes"
	"tst-isp/internal/services"
	"tst-isp/pkg/logger"
)

func validateConfig() error {
	required := []string{"JWT_SECRET", "ROUTER_SECRET_KEY", "MIKROTIK_COA_SECRET"}
	for _, key := range required {
		if val := strings.TrimSpace(os.Getenv(key)); val == "" || val == "change-me" {
			return fmt.Errorf("required env var %s is not set or has default value", key)
		}
	}

	// Validate JWT_SECRET length
	if len(os.Getenv("JWT_SECRET")) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	// Validate ROUTER_SECRET_KEY length
	if len(os.Getenv("ROUTER_SECRET_KEY")) < 32 {
		return fmt.Errorf("ROUTER_SECRET_KEY must be at least 32 characters")
	}

	// Optional but recommended
	if addr := os.Getenv("MIKROTIK_COA_ADDR"); addr != "" {
		if !strings.Contains(addr, ":") {
			return fmt.Errorf("MIKROTIK_COA_ADDR must be in format host:port")
		}
	}

	logger.Info("config validation passed")
	return nil
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		logger.Warn("error loading .env file: %v", err)
	}

	if err := validateConfig(); err != nil {
		logger.Error("config validation failed: %v", err)
		os.Exit(1)
	}

	if _, err := db.InitMongo(); err != nil {
		logger.Error("db error: %v", err)
		os.Exit(1)
	}
	if rdb, err := db.InitRadius(); err != nil {
		// Keep local/dev startup resilient when RADIUS is unavailable.
		// Set RADIUS_REQUIRED=true to enforce a hard dependency.
		if strings.EqualFold(strings.TrimSpace(os.Getenv("RADIUS_REQUIRED")), "true") {
			logger.Error("radius db error: %v", err)
			os.Exit(1)
		}
		logger.Warn("radius db warning: %v (continuing without radius)", err)
	} else if rdb != nil {
		logger.Info("radius db connected")
	} else {
		logger.Info("radius db not configured")
	}

	if err := seedSuperAdmin(); err != nil {
		logger.Error("seed error: %v", err)
		os.Exit(1)
	}
	if err := seedDefaultSubISP(); err != nil {
		logger.Error("seed sub-isp error: %v", err)
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      routes.Register(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	services.StartExpiryWorker(ctx)
	services.StartRadiusReconcileWorker(ctx)

	logger.Info("API listening on :%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server error: %v", err)
		os.Exit(1)
	}
}

func seedSuperAdmin() error {
	email := strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_EMAIL")))
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" || password == "" {
		return nil
	}

	coll := db.DB().Collection("users")
	count, err := coll.CountDocuments(context.Background(), map[string]interface{}{"email": email})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = coll.InsertOne(context.Background(), models.User{
		Email:     email,
		Password:  string(hash),
		Role:      "super_admin",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	return err
}

func seedDefaultSubISP() error {
	email := strings.ToLower(strings.TrimSpace(os.Getenv("SUBISP_EMAIL")))
	password := os.Getenv("SUBISP_PASSWORD")
	name := strings.TrimSpace(os.Getenv("SUBISP_NAME"))
	if email == "" || password == "" || name == "" {
		return nil
	}

	collUsers := db.DB().Collection("users")
	count, err := collUsers.CountDocuments(context.Background(), bson.M{"email": email})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	subID := primitive.NewObjectID().Hex()
	now := time.Now()
	_, err = db.DB().Collection("sub_isps").InsertOne(context.Background(), models.SubISP{
		ID:           subID,
		Name:         name,
		ContactName:  name,
		Email:        email,
		Phone:        "",
		Location:     "",
		RouterCount:  0,
		MaxUsers:     50,
		MaxRouters:   2,
		Routers:      []models.SubRouter{},
		Plan:         "Lite Plan",
		Status:       "pending",
		UsageUsedGB:  0,
		UsageLimitGB: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return err
	}

	_, err = collUsers.InsertOne(context.Background(), models.User{
		Email:     email,
		Password:  string(hash),
		Phone:     "",
		Role:      "sub_isp",
		SubIspID:  subID,
		Package:   "Lite Plan",
		CreatedAt: now,
		UpdatedAt: now,
	})
	return err
}
