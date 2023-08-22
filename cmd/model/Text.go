package model

import (
	"time"
)

type TextMessage struct {
	Name     string    `bson:"name" json:"name"`
	DateTime time.Time `bson:"date_time" json:"date_time"`
	Text     string    `bson:"data" json:"data"`
}
