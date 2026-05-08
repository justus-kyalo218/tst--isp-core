package services

import (
	"context"
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"

	"go.mongodb.org/mongo-driver/bson"

	"tst-isp/internal/db"
	"tst-isp/internal/models"
	"tst-isp/pkg/logger"
	"tst-isp/pkg/utils"
)

// SendDisconnect sends a Disconnect-Request (RFC5176) to MikroTik to force re-auth.
func SendDisconnect(username string) error {
	addr, secret := getCoaConfig()
	if addr == "" || secret == "" {
		logger.Warn("COA not configured, skipping disconnect for user: %s", username)
		return nil
	}
	if username == "" {
		logger.Error("attempted COA disconnect with empty username")
		return errors.New("missing username")
	}

	logger.Info("sending COA disconnect for user: %s to %s", username, addr)

	packet := radius.New(radius.CodeDisconnectRequest, []byte(secret))
	rfc2865.UserName_SetString(packet, username)

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		logger.Error("invalid COA address %s: %v", addr, err)
		return err
	}
	if ip := net.ParseIP(host); ip != nil {
		rfc2865.NASIPAddress_Set(packet, ip)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = radius.Exchange(ctx, packet, addr)
	if err != nil {
		logger.Error("COA disconnect failed for user %s: %v", username, err)
		return err
	}

	logger.Info("COA disconnect successful for user: %s", username)
	return nil
}

func getCoaConfig() (string, string) {
	addr := strings.TrimSpace(os.Getenv("MIKROTIK_COA_ADDR"))
	secret := strings.TrimSpace(os.Getenv("MIKROTIK_COA_SECRET"))
	if addr != "" && secret != "" {
		return addr, secret
	}
	if db.DB() == nil {
		return "", ""
	}

	var rt models.Router
	err := db.DB().Collection("routers").FindOne(context.Background(), bson.M{"enabled": true}).Decode(&rt)
	if err != nil || rt.Host == "" || rt.SecretEnc == "" {
		return "", ""
	}
	sec, err := utils.DecryptSecret(rt.SecretEnc)
	if err != nil {
		return "", ""
	}
	port := rt.CoAPort
	if port == 0 {
		port = 3799
	}
	return net.JoinHostPort(rt.Host, strconv.Itoa(port)), sec
}
