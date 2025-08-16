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
		cmd.warn("delete-group-failed", "cannot delete default group")
		return
	}

	err := cmd.DB.DeleteGroup(data.Id)
	if err != nil {
		log.Printf("there was an error deleting group %s", err.Error())
		cmd.warn("delete-group-failed", "there was an error deleting group")
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
		cmd.warn("add-group-failed", "there was an error adding group")
		return
	}
	cmd.send("group-added", group_added)
}

func (cmd *Cmd) AddProfiles(data structs.AddProfilesData) {
	var profiles_added structs.ProfilesAdded = lib.AddProfiles(cmd.DB, data)
	cmd.send("profiles-added", profiles_added)
}

func (cmd *Cmd) UpdateProfile(data structs.UpdateProfileData) {
	profile_data, err := cmd.DB.RenameProfile(data.Profile, data.Name)
	if err != nil {
		log.Printf("there was an error updating the profile with gid,id: %d,%d: %s", data.Profile.GroupId, data.Profile.Id, err.Error())
		cmd.warn("update-profile-failed", "there was an error updating the profile")
	}
	profile_updated := structs.ProfileUpdated{
		Profile: profile_data,
	}
	cmd.send("profile-updated", profile_updated)
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

func (cmd *Cmd) warn(key string, msg string) {
	cmd.BroadCast(lib.CreateJsonNotification("warn", structs.Warning{Key: key, Content: msg}))
}

func (cmd *Cmd) fatal(msg string) {
	cmd.BroadCast(lib.CreateJsonNotification("error", structs.Warning{Content: msg}))
}

func (cmd *Cmd) send(msg string, obj any) {
	cmd.BroadCast(lib.CreateJsonNotification(msg, obj))
}
