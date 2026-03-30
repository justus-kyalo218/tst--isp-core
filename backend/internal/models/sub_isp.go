package models

import "time"

type SubRouter struct {
	ID     string `bson:"id" json:"id"`
	Name   string `bson:"name" json:"name"`
	Status string `bson:"status" json:"status"`
}

type SubISP struct {
	ID           string    `bson:"_id,omitempty" json:"id"`
	Name         string    `bson:"name" json:"name"`
	ContactName  string    `bson:"contact_name" json:"contactName"`
	Email        string    `bson:"email" json:"email"`
	Phone        string    `bson:"phone" json:"phone"`
	Location     string    `bson:"location" json:"location"`
	RouterCount  int       `bson:"router_count" json:"routerCount"`
	MaxUsers     int       `bson:"max_users" json:"maxUsers"`
	MaxRouters   int       `bson:"max_routers" json:"maxRouters"`
	Routers      []SubRouter  `bson:"routers" json:"routers"`
	Plan         string    `bson:"plan" json:"plan"`
	Status       string    `bson:"status" json:"status"`
	UsageUsedGB  int       `bson:"usage_used_gb" json:"usageUsed"`
	UsageLimitGB int       `bson:"usage_limit_gb" json:"usageLimit"`
	PaidUntil    time.Time `bson:"paid_until" json:"paidUntil"`
	CreatedAt    time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt    time.Time `bson:"updated_at" json:"updatedAt"`
}
