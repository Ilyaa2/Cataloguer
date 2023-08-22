package model

type LinkMessage struct {
	Name string `bson:"name" json:"name"`
	Link string `bson:"link" json:"link"`
}
