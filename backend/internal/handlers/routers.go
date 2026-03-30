package handlers

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"tst-isp/internal/db"
	"tst-isp/pkg/utils"
)

type routerPayload struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	Secret      string `json:"secret"`
	ServiceType string `json:"serviceType"`
	CoAPort     int    `json:"coaPort"`
	AuthPort    int    `json:"authPort"`
	AcctPort    int    `json:"acctPort"`
	NASID       string `json:"nasId"`
	Enabled     bool   `json:"enabled"`
}

type routerResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Host        string    `json:"host"`
	ServiceType string    `json:"serviceType"`
	CoAPort     int       `json:"coaPort"`
	AuthPort    int       `json:"authPort"`
	AcctPort    int       `json:"acctPort"`
	NASID       string    `json:"nasId"`
	Enabled     bool      `json:"enabled"`
	HasSecret   bool      `json:"hasSecret"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func AdminRouters(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listRouters(w, r)
	case http.MethodPost:
		if strings.HasSuffix(r.URL.Path, "/test") {
			testRouter(w, r)
		} else {
			createRouter(w, r)
		}
	case http.MethodPut:
		updateRouter(w, r)
	case http.MethodDelete:
		deleteRouter(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func listRouters(w http.ResponseWriter, r *http.Request) {
	coll := db.DB().Collection("routers")
	cur, err := coll.Find(r.Context(), bson.M{}, options.Find().SetSort(bson.M{"created_at": -1}))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load routers")
		return
	}
	defer cur.Close(r.Context())

	var out []routerResponse
	for cur.Next(r.Context()) {
		var rt struct {
			ID          primitive.ObjectID `bson:"_id"`
			Name        string             `bson:"name"`
			Host        string             `bson:"host"`
			SecretEnc   string             `bson:"secret_enc"`
			ServiceType string             `bson:"service_type"`
			CoAPort     int                `bson:"coa_port"`
			AuthPort    int                `bson:"auth_port"`
			AcctPort    int                `bson:"acct_port"`
			NASID       string             `bson:"nas_id"`
			Enabled     bool               `bson:"enabled"`
			CreatedAt   time.Time          `bson:"created_at"`
			UpdatedAt   time.Time          `bson:"updated_at"`
		}
		if err := cur.Decode(&rt); err != nil {
			continue
		}
		out = append(out, routerResponse{
			ID:          rt.ID.Hex(),
			Name:        rt.Name,
			Host:        rt.Host,
			ServiceType: rt.ServiceType,
			CoAPort:     rt.CoAPort,
			AuthPort:    rt.AuthPort,
			AcctPort:    rt.AcctPort,
			NASID:       rt.NASID,
			Enabled:     rt.Enabled,
			HasSecret:   rt.SecretEnc != "",
			CreatedAt:   rt.CreatedAt,
			UpdatedAt:   rt.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func createRouter(w http.ResponseWriter, r *http.Request) {
	var req routerPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Host = strings.TrimSpace(req.Host)
	if req.Name == "" || req.Host == "" || req.Secret == "" {
		writeError(w, http.StatusBadRequest, "name, host, and secret are required")
		return
	}

	enc, err := utils.EncryptSecret(req.Secret)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	now := time.Now()
	doc := bson.M{
		"name":         req.Name,
		"host":         req.Host,
		"secret_enc":   enc,
		"service_type": req.ServiceType,
		"coa_port":     defaultPort(req.CoAPort, 3799),
		"auth_port":    defaultPort(req.AuthPort, 1812),
		"acct_port":    defaultPort(req.AcctPort, 1813),
		"nas_id":       req.NASID,
		"enabled":      req.Enabled,
		"created_at":   now,
		"updated_at":   now,
	}

	res, err := db.DB().Collection("routers").InsertOne(r.Context(), doc)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save router")
		return
	}

	id := ""
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		id = oid.Hex()
	}

	_ = syncRadiusNAS(r, req.Host, req.Name, req.ServiceType, req.Secret)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func updateRouter(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req routerPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	update := bson.M{
		"$set": bson.M{
			"name":         strings.TrimSpace(req.Name),
			"host":         strings.TrimSpace(req.Host),
			"service_type": req.ServiceType,
			"coa_port":     defaultPort(req.CoAPort, 3799),
			"auth_port":    defaultPort(req.AuthPort, 1812),
			"acct_port":    defaultPort(req.AcctPort, 1813),
			"nas_id":       req.NASID,
			"enabled":      req.Enabled,
			"updated_at":   time.Now(),
		},
	}

	if strings.TrimSpace(req.Secret) != "" {
		enc, err := utils.EncryptSecret(req.Secret)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		update["$set"].(bson.M)["secret_enc"] = enc
	}

	_, err = db.DB().Collection("routers").UpdateOne(r.Context(), bson.M{"_id": oid}, update)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update router")
		return
	}

	if req.Secret != "" && req.Host != "" {
		_ = syncRadiusNAS(r, req.Host, req.Name, req.ServiceType, req.Secret)
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteRouter(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var rt struct {
		Host string `bson:"host"`
	}
	_ = db.DB().Collection("routers").FindOne(r.Context(), bson.M{"_id": oid}).Decode(&rt)
	_, err = db.DB().Collection("routers").DeleteOne(r.Context(), bson.M{"_id": oid})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete router")
		return
	}
	if rt.Host != "" {
		_ = deleteRadiusNAS(r, rt.Host)
	}
	w.WriteHeader(http.StatusNoContent)
}

func defaultPort(v int, def int) int {
	if v > 0 {
		return v
	}
	return def
}

func syncRadiusNAS(r *http.Request, host, name, serviceType, secret string) error {
	rdb, err := db.RadiusDB()
	if err != nil || rdb == nil {
		return nil
	}
	if host == "" || secret == "" {
		return nil
	}
	short := name
	if short == "" {
		short = host
	}
	nasType := "mikrotik"
	if serviceType != "" {
		nasType = serviceType
	}
	_, _ = rdb.ExecContext(r.Context(), "DELETE FROM nas WHERE nasname=?", host)
	_, err = rdb.ExecContext(
		r.Context(),
		"INSERT INTO nas (nasname, shortname, type, secret) VALUES (?,?,?,?)",
		host, short, nasType, secret,
	)
	return err
}

func deleteRadiusNAS(r *http.Request, host string) error {
	rdb, err := db.RadiusDB()
	if err != nil || rdb == nil {
		return nil
	}
	if host == "" {
		return nil
	}
	_, err = rdb.ExecContext(r.Context(), "DELETE FROM nas WHERE nasname=?", host)
	return err
}

func testRouter(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var rt struct {
		Host      string `bson:"host"`
		CoAPort   int    `bson:"coa_port"`
		SecretEnc string `bson:"secret_enc"`
	}
	if err := db.DB().Collection("routers").FindOne(r.Context(), bson.M{"_id": oid}).Decode(&rt); err != nil {
		writeError(w, http.StatusNotFound, "router not found")
		return
	}
	if rt.Host == "" || rt.SecretEnc == "" {
		writeError(w, http.StatusBadRequest, "router missing host/secret")
		return
	}

	coaPort := rt.CoAPort
	if coaPort == 0 {
		coaPort = 3799
	}
	addr := net.JoinHostPort(rt.Host, strconv.Itoa(coaPort))

	conn, err := net.DialTimeout("udp", addr, 2*time.Second)
	if err != nil {
		writeError(w, http.StatusBadRequest, "unable to reach router CoA port")
		return
	}
	_ = conn.Close()

	ok := true
	msg := "router reachable"
	if rdb, err := db.RadiusDB(); err == nil && rdb != nil {
		var count int
		_ = rdb.QueryRowContext(r.Context(), "SELECT COUNT(1) FROM nas WHERE nasname=?", rt.Host).Scan(&count)
		if count == 0 {
			ok = false
			msg = "router reachable, but not in RADIUS nas table"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      ok,
		"message": msg,
	})
}
