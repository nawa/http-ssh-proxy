package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/nawa/http-ssh-proxy/config"
	"github.com/nawa/http-ssh-proxy/proxy"
)

func main() {
	//TODO read config file name from arguments
	cfg, err := config.FromFile("config.yml")
	if err != nil {
		log.Fatal(err)
	}
	httpServer := proxy.NewProxyServer(cfg)
	httpServer.Start()
}
