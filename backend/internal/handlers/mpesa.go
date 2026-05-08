package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"tst-isp/internal/db"
	"tst-isp/internal/services"
)

type stkPushRequest struct {
	Phone       string `json:"phone"`
	Amount      int    `json:"amount"`
	PackageName string `json:"packageName"`
}

type stkPushResponse struct {
	Message           string `json:"message"`
	Phone             string `json:"phone"`
	Amount            int    `json:"amount"`
	Timestamp         string `json:"timestamp"`
	CheckoutRequestID string `json:"checkoutRequestId,omitempty"`
	MerchantRequestID string `json:"merchantRequestId,omitempty"`
}

func MpesaSTKPush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req stkPushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	req.Phone = strings.TrimSpace(req.Phone)
	if len(req.Phone) != 10 || (!strings.HasPrefix(req.Phone, "07") && !strings.HasPrefix(req.Phone, "01")) {
		writeError(w, http.StatusBadRequest, "phone must start with 07 or 01 and be 10 digits")
		return
	}
	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "amount must be greater than zero")
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
		PackageName: req.PackageName,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if resp.CheckoutRequestID != "" {
		_ = storePending(r, resp.CheckoutRequestID, resp.MerchantRequestID, req.Phone, req.Amount, req.PackageName)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stkPushResponse{
		Message:           resp.Message,
		Phone:             resp.Phone,
		Amount:            resp.Amount,
		Timestamp:         resp.Timestamp,
		CheckoutRequestID: resp.CheckoutRequestID,
		MerchantRequestID: resp.MerchantRequestID,
	})
}

func writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": message,
	})
}

func storePending(r *http.Request, checkoutID, merchantID, phone string, amount int, packageName string) error {
	coll := db.DB().Collection("pending_payments")
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"checkout_request_id": checkoutID,
			"merchant_request_id": merchantID,
			"phone":               phone,
			"amount":              amount,
			"package":             packageName,
			"updated_at":          now,
		},
		"$setOnInsert": bson.M{
			"created_at": now,
		},
	}
	_, err := coll.UpdateOne(r.Context(), bson.M{"checkout_request_id": checkoutID}, update, options.Update().SetUpsert(true))
	return err
}
