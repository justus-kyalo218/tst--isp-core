package services

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"tst-isp/internal/db"
	"tst-isp/internal/models"
)

func StartRadiusReconcileWorker(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := reconcileRadiusUsers(ctx); err != nil {
					log.Printf("radius reconcile error: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func reconcileRadiusUsers(ctx context.Context) error {
	if !RadiusEnabled() || db.DB() == nil {
		return nil
	}
	now := time.Now()
	coll := db.DB().Collection("users")
	cur, err := coll.Find(ctx, bson.M{"role": "user"})
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var u models.User
		if err := cur.Decode(&u); err != nil {
			continue
		}
		if u.Phone == "" {
			continue
		}
		if u.PaidUntil.IsZero() || u.PaidUntil.Before(now) {
			_ = RemoveRadiusUser(u.Phone)
			continue
		}
		plan := RadiusPlanForPackage(u.Package)
		_ = UpsertRadiusUser(u.Phone, u.Phone, plan)
	}

	return nil
}
