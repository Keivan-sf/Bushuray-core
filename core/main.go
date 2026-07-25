package main

import (
	"bushuray-core/db"
	"bushuray-core/lib"
	"bushuray-core/lib/TCPServer"
	"bushuray-core/lib/config"
	proxy "bushuray-core/lib/proxy/mainproxy"
	"bushuray-core/structs"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	log.Println("logs: ./core-debug.log")
	log.Println("for unix you can run: tail -f ./core-debug.log")
	log.SetOutput(&lumberjack.Logger{
		Filename:   "core-debug.log",
		MaxSize:    20,
		MaxBackups: 1,
		MaxAge:     0,
		Compress:   false,
	})

	log.SetPrefix("debug: ")
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	stop_sig := make(chan bool, 1)

	appConfig, err := config.LoadAppConfig()
	if err != nil {
		log.Println("failed to load application config:", err, "using defaults")
	}

	database := db.DB{}
	database.Initialize()
	proxy_manager := proxy.ProxyManager{}
	proxy_manager.Init(appConfig)

	server := TCPServer.NewServer(&database, &proxy_manager, stop_sig, appConfig.CoreTCPPort)
	server.Start()

	connectOnStartup(&database, &proxy_manager, appConfig.AutoConnectOnStart)

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

func connectOnStartup(database *db.DB, proxy_manager *proxy.ProxyManager, autoConnect bool) {
	if autoConnect {
		profile, err := database.GetLatestConnectedProfile()
		if err != nil {
			return
		}
		err = proxy_manager.Connect(profile, false)
		if err != nil {
			log.Fatal("failed to connect to profile on startup", err)
		}
	}
}
