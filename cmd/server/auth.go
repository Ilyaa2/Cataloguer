package server

import (
	"net/http"
)

type Auth interface {
	Login(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	Auth(next http.Handler) http.Handler
}
