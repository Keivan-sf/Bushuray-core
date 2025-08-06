package TCPServer

import (
	"bushuray-core/lib"
	"bushuray-core/structs"
	"log"
)

func (s *Server) handleStatusChange() {
	log.Println("listening to connection state change")
	for status := range s.proxy_manager.StatusChanged {
		log.Println("Connection status changed:", status.Connection)
		s.BroadCast(lib.CreateJsonNotification("status-changed", status))
	}
}

func (s *Server) handleTunModeStatusChange() {
	log.Println("listening to tun mode state change")
	for status := range s.tun_namager.StatusChanged {
		log.Println("Tun mode status changed:", status)
		s.BroadCast(lib.CreateJsonNotification("tun-status-changed", structs.TunStatus{IsEnabled: status}))
	}
}
