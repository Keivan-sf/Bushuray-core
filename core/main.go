package main

import (
	"bushuray-core/db"
	"bushuray-core/lib/TCPServer"
	"bushuray-core/lib/proxy"
)

func main() {
	database := db.DB{}
	database.Initialize()
	proxy_manager := proxy.ProxyManager{}
	proxy_manager.Init()
	server := TCPServer.NewServer(&database, &proxy_manager)
	server.Start()
	select {}
}
