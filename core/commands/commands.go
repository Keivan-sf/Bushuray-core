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

func (cmd *Cmd) DeleteProfiles(data structs.DeleteProfilesData) {
	var deleted int
	for _, profile := range data.Profiles {
		err := cmd.DB.DeleteProfile(profile.GroupId, profile.Id)
		if err == nil {
			deleted++
		} else {
			log.Println(err)
		}
	}

	log.Println("delted", deleted, "profiles")
}
