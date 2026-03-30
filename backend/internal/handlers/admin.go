package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"tst-isp/internal/db"
	"tst-isp/internal/models"
)

type adminUser struct {
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role"`
	Package   string    `json:"package"`
	PaidUntil time.Time `json:"paidUntil"`
	Active    bool      `json:"active"`
}

func AdminUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	status := r.URL.Query().Get("status")

	coll := db.DB().Collection("users")
	cur, err := coll.Find(r.Context(), bson.M{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load users")
		return
	}
	defer cur.Close(r.Context())

	now := time.Now()
	var out []adminUser
	for cur.Next(r.Context()) {
		var u models.User
		if err := cur.Decode(&u); err != nil {
			continue
		}
		active := !u.PaidUntil.IsZero() && u.PaidUntil.After(now)
		if status == "active" && !active {
			continue
		}
		if status == "inactive" && active {
			continue
		}
		out = append(out, adminUser{
			Email:     u.Email,
			Phone:     u.Phone,
			Role:      u.Role,
			Package:   u.Package,
			PaidUntil: u.PaidUntil,
			Active:    active,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

type revenueItem struct {
	Package string `json:"package"`
	Total   int    `json:"total"`
	Count   int    `json:"count"`
}

func AdminRevenue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	coll := db.DB().Collection("payments")
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.M{
			"_id":   "$package",
			"total": bson.M{"$sum": "$amount"},
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"total": -1}}},
	}

	cur, err := coll.Aggregate(r.Context(), pipeline)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load revenue")
		return
	}
	defer cur.Close(r.Context())

	var out []revenueItem
	var totalAll int
	var countAll int
	for cur.Next(r.Context()) {
		var row struct {
			ID    string `bson:"_id"`
			Total int    `bson:"total"`
			Count int    `bson:"count"`
		}
		if err := cur.Decode(&row); err != nil {
			continue
		}
		out = append(out, revenueItem{
			Package: row.ID,
			Total:   row.Total,
			Count:   row.Count,
		})
		totalAll += row.Total
		countAll += row.Count
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"items": out,
		"total": totalAll,
		"count": countAll,
	})
}
