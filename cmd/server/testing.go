package server

import "Cataloguer/cmd/model"

func prepareServer() *Server {
	config := NewConfig()
	server := New(*config)
	server.Auth = NewAuth(server)
	server.FunctionalServer.Handler = server.ConfigureRoutes()
	return server
}

func getUser() *model.User {
	return &model.User{
		Email:    "user@example.org",
		Password: "password",
	}
}
