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
	"tst-isp/pkg/utils"
)

// SendDisconnect sends a Disconnect-Request (RFC5176) to MikroTik to force re-auth.
func SendDisconnect(username string) error {
	addr, secret := getCoaConfig()
	if addr == "" || secret == "" {
		return nil
	}
	if username == "" {
		return errors.New("missing username")
	}

	packet := radius.New(radius.CodeDisconnectRequest, []byte(secret))
	rfc2865.UserName_SetString(packet, username)

	host, _, err := net.SplitHostPort(addr)
	if err == nil {
		if ip := net.ParseIP(host); ip != nil {
			rfc2865.NASIPAddress_Set(packet, ip)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = radius.Exchange(ctx, packet, addr)
	return err
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
