package domain

import "time"

type CryptoPolicy struct {
	ID        string    `bson:"_id" json:"id"`
	Version   uint32    `bson:"version" json:"version"`
	KDFParams KDFParams `bson:"kdf_params" json:"kdfParams"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
}

type KDFParams struct {
	Algorithm   string `bson:"algorithm" json:"algorithm"`
	Salt        string `bson:"salt" json:"salt"`
	Memory      uint32 `bson:"memory" json:"memory"`
	Iterations  uint32 `bson:"iterations" json:"iterations"`
	Parallelism uint8  `bson:"parallelism" json:"parallelism"`
}
