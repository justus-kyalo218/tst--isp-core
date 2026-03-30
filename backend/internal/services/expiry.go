package services

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"tst-isp/internal/db"
	"tst-isp/internal/models"
)

func StartExpiryWorker(ctx context.Context) {
	interval := 5 * time.Minute
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := runExpirySweep(ctx); err != nil {
					log.Printf("expiry sweep error: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func runExpirySweep(ctx context.Context) error {
	if db.DB() == nil {
		return nil
	}
	now := time.Now()
	collUsers := db.DB().Collection("users")
	cur, err := collUsers.Find(ctx, bson.M{"paid_until": bson.M{"$lt": now}})
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var u models.User
		if err := cur.Decode(&u); err != nil {
			continue
		}
		if u.Role == "user" && u.Phone != "" && RadiusEnabled() {
			filterID := strings.TrimSpace(os.Getenv("RADIUS_WALLED_GARDEN_FILTER_ID"))
			if filterID != "" {
				_ = SetRadiusWalledGarden(u.Phone, filterID)
			} else {
				_ = RemoveRadiusUser(u.Phone)
			}
			_ = SendDisconnect(u.Phone)
		}
	}

	collSub := db.DB().Collection("sub_isps")
	_, _ = collSub.UpdateMany(ctx, bson.M{
		"status":     "active",
		"paid_until": bson.M{"$lt": now},
	}, bson.M{
		"$set": bson.M{
			"status":     "suspended",
			"updated_at": now,
		},
	})

	return nil
}
