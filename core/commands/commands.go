package cmd

import (
	"bushuray-core/db"
	"bushuray-core/lib"
	"bushuray-core/structs"
	"net"
)

type Cmd struct {
	Conn net.Conn
	DB   *db.DB
}

func (cmd *Cmd) AddProfiles(data structs.AddProfilesData) structs.ProfilesAdded {
	return lib.AddProfiles(cmd.DB, data)
}
