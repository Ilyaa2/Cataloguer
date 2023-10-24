package main

import (
	"Cataloguer/cmd/server"
	"encoding/json"
	"flag"
	"log"
	"os"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config-path", "configs/config.json", "config file path in json format")
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}
	config := server.Config{}
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
	serv := server.New(config)
	serv.Start()
	log.Println("server started")
}
