package db

import (
	"bushuray-core/structs"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func (db *DB) DeleteProfile(group_id int, id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	profile_config_path := db.GetProfileFilePath(group_id, id)
	err := os.Remove(profile_config_path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Profile removal error: %s did not exist in the first place", profile_config_path)
		}
		return fmt.Errorf("Failed to delete config file %w", err)
	}
	return nil
}

func (db *DB) AddProfile(data structs.DBAddProfileData) (structs.ProfileAdded, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var profile_added structs.ProfileAdded
	group_data, err := db.loadGroupConfig(data.GroupId)
	if err != nil {
		return profile_added, err
	}
	group_data.LastId++
	err = db.saveGroupConfig(group_data)
	if err != nil {
		return profile_added, err
	}

	profile_id := group_data.LastId
	profile_path := db.GetProfileFilePath(group_data.Id, profile_id)
	err = os.Remove(db.GetProfileFilePath(group_data.Id, profile_id))
	if err != nil && !os.IsNotExist(err) {
		return profile_added, err
	}
	profile := structs.Profile{
		Id:       profile_id,
		Name:     data.Name,
		Protocol: data.Protocol,
		Uri:      data.Uri,
	}
	profile_json, err := json.MarshalIndent(profile, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(profile_path, profile_json, 0644)
	if err != nil {
		return profile_added, fmt.Errorf("failed to write %s: %w", profile_path, err)
	}

	profile_added = structs.ProfileAdded{
		Uri:      data.Uri,
		GroupId:  data.GroupId,
		Id:       profile_id,
		Protocol: profile.Protocol,
		Name:     profile.Name,
	}

	return profile_added, nil
}
