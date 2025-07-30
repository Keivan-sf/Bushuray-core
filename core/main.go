package main

import (
	"bushuray-core/db"
	"bushuray-core/lib/TCPServer"
	"bushuray-core/lib/proxy"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
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
