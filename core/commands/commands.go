package cmd

import (
	"bushuray-core/db"
	"bushuray-core/lib"
	"bushuray-core/structs"
	"encoding/binary"
	"encoding/json"
	"log"
	"net"
)

type Cmd struct {
	Conn net.Conn
	DB   *db.DB
}

func (cmd *Cmd) AddProfiles(data structs.AddProfilesData) {
	var profiles_added structs.ProfilesAdded = lib.AddProfiles(cmd.DB, data)
	cmd.send("profiles-added", profiles_added)
}

func (cmd *Cmd) DeleteProfiles(data structs.DeleteProfilesData) {
	var deleted structs.ProfilesDeleted
	for _, profile := range data.Profiles {
		err := cmd.DB.DeleteProfile(profile.GroupId, profile.Id)
		if err == nil {
			deleted.DeletedProfiles = append(deleted.DeletedProfiles, profile)
		} else {
			log.Println(err)
		}
	}
	cmd.send("profiles-deleted", deleted)
}

func (cmd *Cmd) send(msg string, obj any) {
	data := structs.Message[any]{
		Msg:  msg,
		Data: obj,
	}
	json_data, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("failed to parse send message %v %s", data, err)
	}
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(json_data)))

	log.Println(string(json_data))

	cmd.Conn.Write(length)
	cmd.Conn.Write(json_data)
}
