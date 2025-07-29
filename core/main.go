package main

import (
	"bushuray-core/db"
	"bushuray-core/lib/TCPServer"
)

func main() {
	database := db.DB{}
	database.Initialize()
	server := TCPServer.NewServer(&database)
	server.Start()
	select {}
}
