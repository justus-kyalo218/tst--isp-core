package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"tst-isp/internal/db"
	"tst-isp/internal/models"
	"tst-isp/internal/services"
)

func MpesaCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	log.Printf("mpesa callback: %v", payload)

	resultCode, checkoutID, packageName, amount, phone, receipt := extractCallback(payload)
	if resultCode == 0 && phone != "" && amount > 0 {
		if err := recordPayment(r, checkoutID, phone, packageName, amount, receipt); err != nil {
			log.Printf("mpesa callback storage error: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func extractCallback(payload map[string]interface{}) (int, string, string, int, string, string) {
	body, _ := payload["Body"].(map[string]interface{})
	stkCallback, _ := body["stkCallback"].(map[string]interface{})

	resultCode := intFromAny(stkCallback["ResultCode"])
	checkoutID, _ := stkCallback["CheckoutRequestID"].(string)
	merchantID, _ := stkCallback["MerchantRequestID"].(string)
	_ = merchantID

	// Preferred: AccountReference in CallbackMetadata if present
	meta, _ := stkCallback["CallbackMetadata"].(map[string]interface{})
	items, _ := meta["Item"].([]interface{})

	var amount int
	var phone string
	var receipt string
	var accountRef string

	for _, it := range items {
		item, _ := it.(map[string]interface{})
		name, _ := item["Name"].(string)
		switch name {
		case "Amount":
			amount = intFromAny(item["Value"])
		case "PhoneNumber":
			phone = normalizePhoneAny(item["Value"])
		case "MpesaReceiptNumber":
			if v, ok := item["Value"].(string); ok {
				receipt = v
			}
		case "AccountReference":
			if v, ok := item["Value"].(string); ok {
				accountRef = v
			}
		}
	}

	packageName := accountRef
	return resultCode, checkoutID, packageName, amount, phone, receipt
}

func recordPayment(r *http.Request, checkoutID, phone, packageName string, amount int, receipt string) error {
	collPayments := db.DB().Collection("payments")
	collUsers := db.DB().Collection("users")
	collPending := db.DB().Collection("pending_payments")

	now := time.Now()
	var pending struct {
		Package  string `bson:"package"`
		Kind     string `bson:"kind"`
		SubIspID string `bson:"sub_isp_id"`
		Email    string `bson:"email"`
		Phone    string `bson:"phone"`
	}
	if packageName == "" && checkoutID != "" {
		_ = collPending.FindOne(r.Context(), bson.M{"checkout_request_id": checkoutID}).Decode(&pending)
	}
	if pending.Package != "" {
		packageName = pending.Package
	}
	if pending.Phone != "" {
		phone = pending.Phone
	}
	if pending.Kind == "sub_isp" && pending.SubIspID != "" {
		if _, err := collPayments.InsertOne(r.Context(), models.Payment{
			Phone:     phone,
			Package:   packageName,
			Amount:    amount,
			Kind:      "sub_isp",
			SubIspID:  pending.SubIspID,
			Ref:       receipt,
			CreatedAt: now,
		}); err != nil {
			return err
		}

		paidUntil := now.Add(subIspPackageDuration(packageName))
		maxUsers, maxRouters := subIspPlanLimits(packageName)
		collSub := db.DB().Collection("sub_isps")
		_, err := collSub.UpdateOne(r.Context(), bson.M{"_id": pending.SubIspID}, bson.M{
			"$set": bson.M{
				"plan":        packageName,
				"status":      "active",
				"paid_until":  paidUntil,
				"max_users":   maxUsers,
				"max_routers": maxRouters,
				"updated_at":  now,
			},
		})
		if err != nil {
			return err
		}

		if pending.Email != "" {
			password := pending.Phone
			if password != "" {
				hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					return err
				}
				update := bson.M{
					"$set": bson.M{
						"package":    packageName,
						"paid_until": paidUntil,
						"updated_at": now,
					},
					"$setOnInsert": bson.M{
						"email":      pending.Email,
						"password":   string(hash),
						"phone":      pending.Phone,
						"role":       "sub_isp",
						"sub_isp_id": pending.SubIspID,
						"created_at": now,
					},
				}
				_, _ = collUsers.UpdateOne(r.Context(), bson.M{"email": pending.Email}, update, options.Update().SetUpsert(true))
			}
		}
		if checkoutID != "" {
			_, _ = collPending.DeleteOne(r.Context(), bson.M{"checkout_request_id": checkoutID})
		} else if pending.Email != "" {
			_, _ = collPending.DeleteMany(r.Context(), bson.M{"email": pending.Email, "kind": "sub_isp"})
		}
		return nil
	}
	if _, err := collPayments.InsertOne(r.Context(), models.Payment{
		Phone:     phone,
		Package:   packageName,
		Amount:    amount,
		Kind:      "user",
		Ref:       receipt,
		CreatedAt: now,
	}); err != nil {
		return err
	}

	paidUntil := now.Add(packageDuration(packageName))
	update := bson.M{
		"$set": bson.M{
			"phone":      phone,
			"package":    packageName,
			"paid_until": paidUntil,
			"updated_at": now,
		},
		"$setOnInsert": bson.M{
			"role":       "user",
			"created_at": now,
		},
	}
	_, err := collUsers.UpdateOne(r.Context(), bson.M{"phone": phone}, update, options.Update().SetUpsert(true))
	if err == nil && services.RadiusEnabled() {
		username := phone
		password := phone
		plan := services.RadiusPlanForPackage(packageName)
		_ = services.UpsertRadiusUser(username, password, plan)
		_ = services.SendDisconnect(username)
	}
	return err
}

func intFromAny(v interface{}) int {
	switch t := v.(type) {
	case int:
		return t
	case int32:
		return int(t)
	case int64:
		return int(t)
	case float64:
		return int(t)
	default:
		return 0
	}
}

func normalizePhoneAny(v interface{}) string {
	s, _ := v.(string)
	if s == "" {
		// sometimes number is numeric
		if n := intFromAny(v); n != 0 {
			s = fmt.Sprintf("%d", n)
		}
	}
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "+")
	if strings.HasPrefix(s, "07") && len(s) == 10 {
		return "254" + s[1:]
	}
	if strings.HasPrefix(s, "2547") && len(s) == 12 {
		return s
	}
	return s
}

func packageDuration(name string) time.Duration {
	n := strings.ToLower(strings.TrimSpace(name))
	switch {
	case strings.HasPrefix(n, "sub-isp"):
		return 30 * 24 * time.Hour
	case strings.HasPrefix(n, "30 mins"):
		return 30 * time.Minute
	case strings.HasPrefix(n, "1 hour"):
		return 1 * time.Hour
	case strings.HasPrefix(n, "2 hours"):
		return 2 * time.Hour
	case strings.HasPrefix(n, "6 hours"):
		return 6 * time.Hour
	case strings.HasPrefix(n, "12 hours"):
		return 12 * time.Hour
	case strings.HasPrefix(n, "24 hours"):
		return 24 * time.Hour
	case strings.HasPrefix(n, "3 days"):
		return 72 * time.Hour
	case strings.HasPrefix(n, "7 days"):
		return 7 * 24 * time.Hour
	case strings.HasPrefix(n, "15 days"):
		return 15 * 24 * time.Hour
	case strings.HasPrefix(n, "30 days"):
		return 30 * 24 * time.Hour
	default:
		return 24 * time.Hour
	}
}

func subIspPackageDuration(_ string) time.Duration {
	return 30 * 24 * time.Hour
}
