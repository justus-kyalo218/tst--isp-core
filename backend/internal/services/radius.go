package services

import (
	"database/sql"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
	"sync"

	"tst-isp/internal/db"
)

type RadiusPlan struct {
	RateLimit      string
	SessionTimeout int
	TotalLimit     int
}

func RadiusEnabled() bool {
	_, err := db.RadiusDB()
	return err == nil
}

func planRateLimit(packageName string) string {
	n := strings.ToLower(packageName)
	switch {
	case strings.Contains(n, "10 mbps"):
		return "10M/10M"
	case strings.Contains(n, "8 mbps"):
		return "8M/8M"
	case strings.Contains(n, "5 mbps"):
		return "5M/5M"
	case strings.Contains(n, "4 mbps"):
		return "4M/4M"
	default:
		return "4M/4M"
	}
}

func RadiusPlanForPackage(packageName string) RadiusPlan {
	duration := packageDurationSeconds(packageName)
	return RadiusPlan{
		RateLimit:      planRateLimit(packageName),
		SessionTimeout: duration,
		TotalLimit:     planTotalLimitBytes(packageName),
	}
}

var (
	totalLimitOnce sync.Once
	totalLimitMap  map[string]int
)

func planTotalLimitBytes(packageName string) int {
	totalLimitOnce.Do(func() {
		raw := strings.TrimSpace(os.Getenv("RADIUS_TOTAL_LIMITS"))
		if raw != "" {
			totalLimitMap = parseTotalLimitEnv(raw)
		} else {
			totalLimitMap = defaultTotalLimits()
		}
	})
	key := strings.ToLower(strings.TrimSpace(packageName))
	if v, ok := totalLimitMap[key]; ok {
		return v
	}
	return 0
}

func defaultTotalLimits() map[string]int {
	return map[string]int{
		"30 mins 4 mbps":  300 * 1024 * 1024,
		"1 hour 4 mbps":   500 * 1024 * 1024,
		"2 hours 4 mbps":  1024 * 1024 * 1024,
		"6 hours 4 mbps":  2048 * 1024 * 1024,
		"12 hours 4 mbps": 3072 * 1024 * 1024,
		"24 hours 4 mbps": 4096 * 1024 * 1024,
		"3 days 4 mbps":   8192 * 1024 * 1024,
		"7 days 5 mbps":   15360 * 1024 * 1024,
		"15 days 5 mbps":  30720 * 1024 * 1024,
		"30 days 5 mbps":  51200 * 1024 * 1024,
		"30 days 8 mbps":  81920 * 1024 * 1024,
		"30 days 10 mbps": 122880 * 1024 * 1024,
	}
}

// RADIUS_TOTAL_LIMITS format:
// "30 mins=500MB,1 hour=1GB,2 hours=2GB"
func parseTotalLimitEnv(raw string) map[string]int {
	out := map[string]int{}
	if raw == "" {
		return out
	}
	parts := strings.Split(raw, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(kv[0]))
		val := strings.TrimSpace(kv[1])
		if key == "" || val == "" {
			continue
		}
		if bytes := parseBytes(val); bytes > 0 {
			out[key] = bytes
		}
	}
	return out
}

func parseBytes(v string) int {
	v = strings.ToUpper(strings.TrimSpace(v))
	if v == "" {
		return 0
	}
	multiplier := 1
	switch {
	case strings.HasSuffix(v, "GB"):
		multiplier = 1024 * 1024 * 1024
		v = strings.TrimSpace(strings.TrimSuffix(v, "GB"))
	case strings.HasSuffix(v, "MB"):
		multiplier = 1024 * 1024
		v = strings.TrimSpace(strings.TrimSuffix(v, "MB"))
	case strings.HasSuffix(v, "KB"):
		multiplier = 1024
		v = strings.TrimSpace(strings.TrimSuffix(v, "KB"))
	case strings.HasSuffix(v, "B"):
		multiplier = 1
		v = strings.TrimSpace(strings.TrimSuffix(v, "B"))
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil || n <= 0 {
		return 0
	}
	return int(n * float64(multiplier))
}

func packageDurationSeconds(name string) int {
	n := strings.ToLower(strings.TrimSpace(name))
	switch {
	case strings.HasPrefix(n, "sub-isp"):
		return int((30 * 24 * time.Hour).Seconds())
	case strings.HasPrefix(n, "30 mins"):
		return int((30 * time.Minute).Seconds())
	case strings.HasPrefix(n, "1 hour"):
		return int(time.Hour.Seconds())
	case strings.HasPrefix(n, "2 hours"):
		return int((2 * time.Hour).Seconds())
	case strings.HasPrefix(n, "6 hours"):
		return int((6 * time.Hour).Seconds())
	case strings.HasPrefix(n, "12 hours"):
		return int((12 * time.Hour).Seconds())
	case strings.HasPrefix(n, "24 hours"):
		return int((24 * time.Hour).Seconds())
	case strings.HasPrefix(n, "3 days"):
		return int((72 * time.Hour).Seconds())
	case strings.HasPrefix(n, "7 days"):
		return int((7 * 24 * time.Hour).Seconds())
	case strings.HasPrefix(n, "15 days"):
		return int((15 * 24 * time.Hour).Seconds())
	case strings.HasPrefix(n, "30 days"):
		return int((30 * 24 * time.Hour).Seconds())
	default:
		return int((24 * time.Hour).Seconds())
	}
}

func UpsertRadiusUser(username, password string, plan RadiusPlan) error {
	rdb, err := db.RadiusDB()
	if err != nil {
		return err
	}
	if username == "" || password == "" {
		return errors.New("missing username or password")
	}

	tx, err := rdb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := deleteRadiusUser(tx, username); err != nil {
		return err
	}

	if _, err := tx.Exec(
		"INSERT INTO radcheck (username, attribute, op, value) VALUES (?, 'Cleartext-Password', ':=', ?)",
		username, password,
	); err != nil {
		return err
	}

	if plan.RateLimit != "" {
		if _, err := tx.Exec(
			"INSERT INTO radreply (username, attribute, op, value) VALUES (?, 'Mikrotik-Rate-Limit', ':=', ?)",
			username, plan.RateLimit,
		); err != nil {
			return err
		}
	}
	if plan.SessionTimeout > 0 {
		if _, err := tx.Exec(
			"INSERT INTO radreply (username, attribute, op, value) VALUES (?, 'Session-Timeout', ':=', ?)",
			username, plan.SessionTimeout,
		); err != nil {
			return err
		}
	}
	if plan.TotalLimit > 0 {
		if _, err := tx.Exec(
			"INSERT INTO radreply (username, attribute, op, value) VALUES (?, 'Mikrotik-Total-Limit', ':=', ?)",
			username, plan.TotalLimit,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func RemoveRadiusUser(username string) error {
	rdb, err := db.RadiusDB()
	if err != nil {
		return err
	}
	_, _ = rdb.Exec("DELETE FROM radreply WHERE username=?", username)
	_, _ = rdb.Exec("DELETE FROM radcheck WHERE username=?", username)
	return nil
}

func SetRadiusWalledGarden(username, filterID string) error {
	rdb, err := db.RadiusDB()
	if err != nil {
		return err
	}
	_, _ = rdb.Exec("DELETE FROM radreply WHERE username=?", username)
	_, err = rdb.Exec(
		"INSERT INTO radreply (username, attribute, op, value) VALUES (?, 'Filter-Id', ':=', ?)",
		username, filterID,
	)
	return err
}

func deleteRadiusUser(tx *sql.Tx, username string) error {
	if _, err := tx.Exec("DELETE FROM radreply WHERE username=?", username); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM radcheck WHERE username=?", username); err != nil {
		return err
	}
	return nil
}
