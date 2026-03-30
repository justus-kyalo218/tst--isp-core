package models

import "time"

type Payment struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Phone     string    `bson:"phone" json:"phone"`
	Package   string    `bson:"package" json:"package"`
	Amount    int       `bson:"amount" json:"amount"`
	Kind      string    `bson:"kind,omitempty" json:"kind,omitempty"`
	SubIspID  string    `bson:"sub_isp_id,omitempty" json:"subIspId,omitempty"`
	Ref       string    `bson:"ref" json:"ref"`
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
}
