package TCPServer

import (
	"bushuray-core/lib"
	"log"
)

func (s *Server) handleStatusChange() {
	log.Println("listening to connection state change")
	for status := range s.proxy_manager.StatusChanged {
		s.BroadCast(lib.CreateJsonNotification("status-changed", status))
	}
}
