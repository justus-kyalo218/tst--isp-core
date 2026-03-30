package models

import "time"

type Router struct {
	ID          string    `bson:"_id,omitempty" json:"id"`
	Name        string    `bson:"name" json:"name"`
	Host        string    `bson:"host" json:"host"`
	SecretEnc   string    `bson:"secret_enc" json:"-"`
	ServiceType string    `bson:"service_type" json:"serviceType"`
	CoAPort     int       `bson:"coa_port" json:"coaPort"`
	AuthPort    int       `bson:"auth_port" json:"authPort"`
	AcctPort    int       `bson:"acct_port" json:"acctPort"`
	NASID       string    `bson:"nas_id" json:"nasId"`
	Enabled     bool      `bson:"enabled" json:"enabled"`
	CreatedAt   time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updatedAt"`
}
