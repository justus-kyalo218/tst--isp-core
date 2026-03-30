package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"tst-isp/internal/db"
)

func SubIspUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	username := strings.TrimSpace(r.URL.Query().Get("username"))
	if username == "" {
		writeError(w, http.StatusBadRequest, "username required")
		return
	}

	rdb, err := db.RadiusDB()
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "radius db not configured")
		return
	}

	row := rdb.QueryRow(
		"SELECT COALESCE(SUM(acctinputoctets + acctoutputoctets),0) AS bytesUsed, COALESCE(SUM(acctsessiontime),0) AS timeUsed FROM radacct WHERE username=?",
		username,
	)

	var bytesUsed int64
	var timeUsed int64
	if err := row.Scan(&bytesUsed, &timeUsed); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load usage")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(usageResponse{
		Username:  username,
		BytesUsed: bytesUsed,
		TimeUsed:  timeUsed,
	})
}
