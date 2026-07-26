package TCPServer

import (
	"bushuray-core/lib"
	"bushuray-core/structs"
	"log"
)

func (s *Server) handleStatusChange() {
	log.Println("listening to connection state change")
	for status := range s.proxy_manager.StatusChanged {
		log.Printf(
			"publishing proxy status: connection=%s tun=%t profile=%d/%d",
			status.Connection,
			status.IsTunEnabled,
			status.Profile.GroupId,
			status.Profile.Id,
		)
		s.BroadCast(lib.CreateJsonNotification("status-changed", status))
		s.BroadCast(lib.CreateJsonNotification("tun-status-changed", structs.TunStatus{
			IsEnabled: status.IsTunEnabled,
		}))
	}
}
