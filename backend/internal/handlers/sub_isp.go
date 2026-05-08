package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"tst-isp/internal/db"
	"tst-isp/internal/middleware"
	"tst-isp/internal/models"
	"tst-isp/internal/services"
)

type subIspRegisterRequest struct {
	Business    string `json:"business"`
	Contact     string `json:"contact"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Phone       string `json:"phone"`
	Location    string `json:"location"`
	PackageName string `json:"packageName"`
	Amount      int    `json:"amount"`
}

type subIspRegisterResponse struct {
	Message           string `json:"message"`
	Phone             string `json:"phone"`
	Amount            int    `json:"amount"`
	Timestamp         string `json:"timestamp"`
	CheckoutRequestID string `json:"checkoutRequestId,omitempty"`
	MerchantRequestID string `json:"merchantRequestId,omitempty"`
	Token             string `json:"token,omitempty"`
	Role              string `json:"role,omitempty"`
}

func SubIspRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req subIspRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	req.Business = strings.TrimSpace(req.Business)
	req.Contact = strings.TrimSpace(req.Contact)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Phone = strings.TrimSpace(req.Phone)
	req.Location = strings.TrimSpace(req.Location)
	req.PackageName = strings.TrimSpace(req.PackageName)

	if req.Business == "" || req.Email == "" {
		writeError(w, http.StatusBadRequest, "business name and email are required")
		return
	}
	if req.Password != "" && len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if len(req.Phone) != 10 || (!strings.HasPrefix(req.Phone, "07") && !strings.HasPrefix(req.Phone, "01")) {
		writeError(w, http.StatusBadRequest, "phone must start with 07 or 01 and be 10 digits")
		return
	}
	if req.PackageName == "" {
		writeError(w, http.StatusBadRequest, "package is required")
		return
	}
	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "amount must be greater than zero")
		return
	}

	collUsers := db.DB().Collection("users")
	if count, _ := collUsers.CountDocuments(r.Context(), bson.M{"email": req.Email}); count > 0 {
		writeError(w, http.StatusConflict, "email already registered")
		return
	}
	collSubIsps := db.DB().Collection("sub_isps")

	now := time.Now()
	maxUsers, maxRouters := subIspPlanLimits(req.PackageName)
	var existing models.SubISP
	subIspID := ""
	err := collSubIsps.FindOne(r.Context(), bson.M{"email": req.Email}).Decode(&existing)
	if err == nil {
		if existing.Status != "pending" {
			writeError(w, http.StatusConflict, "sub-isp already registered")
			return
		}
		subIspID = existing.ID
		_, err = collSubIsps.UpdateOne(r.Context(), bson.M{"_id": subIspID}, bson.M{
			"$set": bson.M{
				"name":         req.Business,
				"contact_name": req.Contact,
				"phone":        req.Phone,
				"location":     req.Location,
				"plan":         req.PackageName,
				"status":       "pending",
				"max_users":    maxUsers,
				"max_routers":  maxRouters,
				"updated_at":   now,
			},
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update sub-isp")
			return
		}
	} else if err == mongo.ErrNoDocuments {
		subIspID = primitive.NewObjectID().Hex()
		subIsp := models.SubISP{
			ID:           subIspID,
			Name:         req.Business,
			ContactName:  req.Contact,
			Email:        req.Email,
			Phone:        req.Phone,
			Location:     req.Location,
			RouterCount:  0,
			MaxUsers:     maxUsers,
			MaxRouters:   maxRouters,
			Routers:      []models.SubRouter{},
			Plan:         req.PackageName,
			Status:       "pending",
			UsageUsedGB:  0,
			UsageLimitGB: 0,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if _, err := collSubIsps.InsertOne(r.Context(), subIsp); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create sub-isp")
			return
		}
	} else {
		writeError(w, http.StatusInternalServerError, "failed to check sub-isp")
		return
	}

	client, err := services.NewDarajaFromEnv()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := client.ResolveCallbackURL(r); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := client.STKPush(services.STKPushRequest{
		Phone:       req.Phone,
		Amount:      req.Amount,
		PackageName: "Sub-ISP " + req.PackageName,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if resp.CheckoutRequestID != "" {
		_ = storePendingSubIsp(r, resp.CheckoutRequestID, resp.MerchantRequestID, req.Phone, req.Amount, req.PackageName, subIspID, req.Email)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(subIspRegisterResponse{
		Message:           resp.Message,
		Phone:             resp.Phone,
		Amount:            resp.Amount,
		Timestamp:         resp.Timestamp,
		CheckoutRequestID: resp.CheckoutRequestID,
		MerchantRequestID: resp.MerchantRequestID,
	})
}

type subIspMeResponse struct {
	models.SubISP
}

func SubIspMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	email := middleware.EmailFromContext(r.Context())
	if email == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var user models.User
	if err := db.DB().Collection("users").FindOne(r.Context(), bson.M{"email": email}).Decode(&user); err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if user.SubIspID == "" {
		writeError(w, http.StatusForbidden, "sub-isp not linked")
		return
	}

	var subIsp models.SubISP
	if err := db.DB().Collection("sub_isps").FindOne(r.Context(), bson.M{"_id": user.SubIspID}).Decode(&subIsp); err != nil {
		writeError(w, http.StatusNotFound, "sub-isp not found")
		return
	}
	if !subIsp.PaidUntil.IsZero() && subIsp.PaidUntil.Before(time.Now()) && subIsp.Status == "active" {
		subIsp.Status = "suspended"
		subIsp.UpdatedAt = time.Now()
		_, _ = db.DB().Collection("sub_isps").UpdateOne(r.Context(), bson.M{"_id": user.SubIspID}, bson.M{
			"$set": bson.M{"status": subIsp.Status, "updated_at": subIsp.UpdatedAt},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(subIspMeResponse{SubISP: subIsp})
}

type routerRequest struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func SubIspRouters(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		SubIspAddRouter(w, r)
	case http.MethodPut:
		SubIspUpdateRouter(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func SubIspAddRouter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	email := middleware.EmailFromContext(r.Context())
	if email == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req routerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "router name required")
		return
	}

	subID, err := subIspIDByEmail(r, email)
	if err != nil {
		writeError(w, http.StatusForbidden, "sub-isp not linked")
		return
	}

	var sub models.SubISP
	if err := db.DB().Collection("sub_isps").FindOne(r.Context(), bson.M{"_id": subID}).Decode(&sub); err != nil {
		writeError(w, http.StatusNotFound, "sub-isp not found")
		return
	}
	if sub.MaxRouters >= 0 && sub.RouterCount >= sub.MaxRouters {
		writeError(w, http.StatusBadRequest, "router limit reached")
		return
	}

	router := models.SubRouter{
		ID:     "rt-" + primitive.NewObjectID().Hex(),
		Name:   req.Name,
		Status: "online",
	}

	update := bson.M{
		"$push": bson.M{"routers": router},
		"$inc":  bson.M{"router_count": 1},
		"$set":  bson.M{"updated_at": time.Now()},
	}
	_, err = db.DB().Collection("sub_isps").UpdateOne(r.Context(), bson.M{"_id": subID}, update)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add router")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(router)
}

func SubIspUpdateRouter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	email := middleware.EmailFromContext(r.Context())
	if email == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req routerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.ID = strings.TrimSpace(req.ID)
	req.Status = strings.TrimSpace(req.Status)
	if req.ID == "" || req.Status == "" {
		writeError(w, http.StatusBadRequest, "router id and status required")
		return
	}

	subID, err := subIspIDByEmail(r, email)
	if err != nil {
		writeError(w, http.StatusForbidden, "sub-isp not linked")
		return
	}

	update := bson.M{
		"$set": bson.M{
			"routers.$.status": req.Status,
			"updated_at":       time.Now(),
		},
	}
	filter := bson.M{"_id": subID, "routers.id": req.ID}
	res, err := db.DB().Collection("sub_isps").UpdateOne(r.Context(), filter, update)
	if err != nil || res.MatchedCount == 0 {
		writeError(w, http.StatusBadRequest, "router not found")
		return
	}

	var subIsp models.SubISP
	if err := db.DB().Collection("sub_isps").FindOne(r.Context(), bson.M{"_id": subID}).Decode(&subIsp); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load router")
		return
	}
	for _, router := range subIsp.Routers {
		if router.ID == req.ID {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(router)
			return
		}
	}
	writeError(w, http.StatusBadRequest, "router not found")
}

func subIspIDByEmail(r *http.Request, email string) (string, error) {
	var user models.User
	if err := db.DB().Collection("users").FindOne(r.Context(), bson.M{"email": email}).Decode(&user); err != nil {
		return "", err
	}
	if user.SubIspID == "" {
		return "", mongo.ErrNoDocuments
	}
	return user.SubIspID, nil
}

func storePendingSubIsp(r *http.Request, checkoutID, merchantID, phone string, amount int, packageName, subIspID, email string) error {
	coll := db.DB().Collection("pending_payments")
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"checkout_request_id": checkoutID,
			"merchant_request_id": merchantID,
			"phone":               phone,
			"amount":              amount,
			"package":             packageName,
			"kind":                "sub_isp",
			"sub_isp_id":          subIspID,
			"email":               email,
			"updated_at":          now,
		},
		"$setOnInsert": bson.M{
			"created_at": now,
		},
	}
	_, err := coll.UpdateOne(r.Context(), bson.M{"checkout_request_id": checkoutID}, update, options.Update().SetUpsert(true))
	return err
}

func subIspUsageLimit(plan string) int {
	_, _ = subIspPlanLimits(plan)
	return 0
}

func subIspPlanLimits(plan string) (int, int) {
	n := strings.ToLower(strings.TrimSpace(plan))
	switch {
	case strings.Contains(n, "lite"):
		return 50, 2
	case strings.Contains(n, "pro"):
		return 150, 5
	case strings.Contains(n, "elite"):
		return 400, 10
	case strings.Contains(n, "unlimited"):
		return -1, -1
	default:
		return 50, 2
	}
}
