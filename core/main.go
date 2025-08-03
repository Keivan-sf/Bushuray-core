package main

import (
	"bushuray-core/db"
	"bushuray-core/lib"
	"bushuray-core/lib/AppConfig"
	"bushuray-core/lib/TCPServer"
	proxy "bushuray-core/lib/proxy/mainproxy"
	"bushuray-core/structs"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	stop_sig := make(chan bool, 1)
	appconfig.LoadConfig()
	database := db.DB{}
	database.Initialize()
	proxy_manager := proxy.ProxyManager{}
	proxy_manager.Init()
	server := TCPServer.NewServer(&database, &proxy_manager, stop_sig)
	server.Start()
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		reason := ""
		select {
		case sig := <-sigs:
			reason = fmt.Sprintf("Received signal %v , cleaning up...", sig)
		case <-stop_sig:
			reason = "Received stop request , cleaning up..."
		}
		log.Println(reason)
		proxy_manager.Stop()
		server.BroadCast(lib.CreateJsonNotification("warn", structs.Warning{Key: "died", Content: reason}))
		os.Exit(0)
	}()
	select {}
}
