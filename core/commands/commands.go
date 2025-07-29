package cmd

import (
	"bushuray-core/db"
	"bushuray-core/lib"
	"bushuray-core/structs"
	"log"
	"net"
)

type Cmd struct {
	Conn net.Conn
	DB   *db.DB
}

func (cmd *Cmd) AddProfiles(data structs.AddProfilesData) structs.ProfilesAdded {
	return lib.AddProfiles(cmd.DB, data)
}

func (cmd *Cmd) DeleteProfiles(data structs.DeleteProfilesData) structs.ProfilesDeleted {
	var deleted structs.ProfilesDeleted
	for _, profile := range data.Profiles {
		err := cmd.DB.DeleteProfile(profile.GroupId, profile.Id)
		if err == nil {
			deleted.DeletedProfiles = append(deleted.DeletedProfiles, profile)
		} else {
			log.Println(err)
		}
	}
	return deleted
}
