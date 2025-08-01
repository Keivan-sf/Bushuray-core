package main

import (
	"bushuray-core/db"
	"bushuray-core/lib/AppConfig"
	"bushuray-core/lib/TCPServer"
	proxy "bushuray-core/lib/proxy/mainproxy"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	appconfig.LoadConfig()
	database := db.DB{}
	database.Initialize()
	proxy_manager := proxy.ProxyManager{}
	proxy_manager.Init()
	server := TCPServer.NewServer(&database, &proxy_manager)
	server.Start()
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigs
		log.Printf("Recved signal %v , cleaning up...", sig)
		proxy_manager.Stop()
		os.Exit(0)
	}()
	select {}
}
