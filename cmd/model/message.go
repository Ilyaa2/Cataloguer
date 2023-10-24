package model

import "time"

type Message struct {
	Id       int       `json:"id"`
	Type     string    `json:"type"`
	Name     string    `json:"name"`
	DateTime time.Time `json:"time"`
	Path     string    `json:"path"`
}

type MessagesResponse struct {
	BaseUrl  string     `json:"baseUrl"`
	Messages []*Message `json:"messages"`
}
