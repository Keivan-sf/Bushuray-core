package cmd

import (
	"bushuray-core/db"
	"bushuray-core/lib"
	"bushuray-core/structs"
	"log"
	"net"
)

type Cmd struct {
	Conn      net.Conn
	BroadCast func([]byte)
	DB        *db.DB
}

func (cmd *Cmd) DeleteGroup(data structs.DeleteGroupData) {
	if data.Id == 0 {
		// cannot delete default group
		return
	}

	err := cmd.DB.DeleteGroup(data.Id)
	if err != nil {
		log.Printf("there was an error adding group %s", err.Error())
		// handle err , warn clients
		return
	}
	group_deleted := structs.GroupDeleted{
		Id: data.Id,
	}
	cmd.send("group-deleted", group_deleted)
}

func (cmd *Cmd) AddGroup(data structs.AddGroupData) {
	group_added, err := cmd.DB.AddGroup(data.Name, data.SubscriptionUrl)
	if err != nil {
		log.Printf("there was an error adding group %s", err.Error())
		// handle err
		return
	}
	cmd.send("group-added", group_added)
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
	cmd.BroadCast(lib.CreateJsonNotification(msg, obj))
}
