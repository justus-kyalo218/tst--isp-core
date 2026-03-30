package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"tst-isp/internal/db"
	"tst-isp/internal/models"
)

type adminSubIsp struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ContactName string    `json:"contactName"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Plan        string    `json:"plan"`
	Status      string    `json:"status"`
	UsageUsed   int       `json:"usageUsed"`
	UsageLimit  int       `json:"usageLimit"`
	RouterCount int       `json:"routerCount"`
	MaxUsers    int       `json:"maxUsers"`
	MaxRouters  int       `json:"maxRouters"`
	PaidUntil   time.Time `json:"paidUntil"`
}

func AdminSubIsps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	status := r.URL.Query().Get("status")
	coll := db.DB().Collection("sub_isps")
	cur, err := coll.Find(r.Context(), bson.M{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load sub-isps")
		return
	}
	defer cur.Close(r.Context())

	var out []adminSubIsp
	for cur.Next(r.Context()) {
		var sub models.SubISP
		if err := cur.Decode(&sub); err != nil {
			continue
		}
		if status != "" && status != "all" && sub.Status != status {
			continue
		}
		out = append(out, adminSubIsp{
			ID:          sub.ID,
			Name:        sub.Name,
			ContactName: sub.ContactName,
			Email:       sub.Email,
			Phone:       sub.Phone,
			Plan:        sub.Plan,
			Status:      sub.Status,
			UsageUsed:   sub.UsageUsedGB,
			UsageLimit:  sub.UsageLimitGB,
			RouterCount: sub.RouterCount,
			MaxUsers:    sub.MaxUsers,
			MaxRouters:  sub.MaxRouters,
			PaidUntil:   sub.PaidUntil,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

type adminSubIspUpdateRequest struct {
	ID     string `json:"id"`
	Plan   string `json:"plan"`
	Status string `json:"status"`
}

func AdminUpdateSubIsp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req adminSubIspUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.ID = strings.TrimSpace(req.ID)
	req.Plan = strings.TrimSpace(req.Plan)
	req.Status = strings.TrimSpace(req.Status)
	if req.ID == "" {
		writeError(w, http.StatusBadRequest, "id required")
		return
	}

	update := bson.M{}
	if req.Plan != "" {
		maxUsers, maxRouters := subIspPlanLimits(req.Plan)
		update["plan"] = req.Plan
		update["max_users"] = maxUsers
		update["max_routers"] = maxRouters
	}
	if req.Status != "" {
		update["status"] = req.Status
	}
	if len(update) == 0 {
		writeError(w, http.StatusBadRequest, "nothing to update")
		return
	}
	update["updated_at"] = time.Now()

	coll := db.DB().Collection("sub_isps")
	_, err := coll.UpdateOne(r.Context(), bson.M{"_id": req.ID}, bson.M{"$set": update})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update sub-isp")
		return
	}

	var sub models.SubISP
	if err := coll.FindOne(r.Context(), bson.M{"_id": req.ID}).Decode(&sub); err != nil {
		writeError(w, http.StatusNotFound, "sub-isp not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(adminSubIsp{
		ID:          sub.ID,
		Name:        sub.Name,
		ContactName: sub.ContactName,
		Email:       sub.Email,
		Phone:       sub.Phone,
		Plan:        sub.Plan,
		Status:      sub.Status,
		UsageUsed:   sub.UsageUsedGB,
		UsageLimit:  sub.UsageLimitGB,
		RouterCount: sub.RouterCount,
		MaxUsers:    sub.MaxUsers,
		MaxRouters:  sub.MaxRouters,
		PaidUntil:   sub.PaidUntil,
	})
}
