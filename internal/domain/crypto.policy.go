package domain

import "time"

type CryptoPolicy struct {
	ID        string    `bson:"_id" json:"id"`
	Version   uint32    `bson:"version" json:"version"`
	KDFParams KDFParams `bson:"kdf_params" json:"kdfParams"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
}
