package models

import "time"

type User struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Email     string    `bson:"email" json:"email"`
	Password  string    `bson:"password" json:"-"`
	Phone     string    `bson:"phone" json:"phone"`
	Role      string    `bson:"role" json:"role"`
	SubIspID  string    `bson:"sub_isp_id,omitempty" json:"subIspId,omitempty"`
	Package   string    `bson:"package" json:"package"`
	PaidUntil time.Time `bson:"paid_until" json:"paidUntil"`
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
}
