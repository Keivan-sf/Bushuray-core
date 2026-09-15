package TCPServer

import (
	"bushuray-core/lib"
	"bushuray-core/structs"
	"log"
)

func (s *Server) handleTestResults() {
	log.Println("listening to test results")
	for result := range s.proxy_manager.TestResultChannel {
		err := s.DB.UpdateProfile(result.Profile)
		if err != nil {
			continue
		}
		if s.removeFailedProfiles && result.Profile.TestResult == -1 {
			status := s.proxy_manager.GetStatus()
			is_connected := status.Connection == "connected" && status.Profile.GroupId == result.Profile.GroupId && status.Profile.Id == result.Profile.Id
			if !is_connected {
				err := s.DB.DeleteProfile(result.Profile.GroupId, result.Profile.Id)
				if err == nil {
					profiles_deleted := structs.ProfilesDeleted{
						DeletedProfiles: []structs.ProfileID{{Id: result.Profile.Id, GroupId: result.Profile.GroupId}},
					}
					s.BroadCast(lib.CreateJsonNotification("profiles-deleted", profiles_deleted))
					continue
				}
				log.Println(err)
			}
		}
		profile_updated := structs.ProfileUpdated{
			Profile: result.Profile,
		}
		s.BroadCast(lib.CreateJsonNotification("profile-updated", profile_updated))
	}
}
