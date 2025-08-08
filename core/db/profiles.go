package db

import (
	"bushuray-core/structs"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"syscall"
)

func (db *DB) UpdateProfile(profile structs.Profile) error {
	oldUmask := syscall.Umask(0)
	defer syscall.Umask(oldUmask)

	db.mu.Lock()
	defer db.mu.Unlock()
	profile_config_path := db.GetProfileFilePath(profile.GroupId, profile.Id)
	_, err := os.ReadFile(profile_config_path)
	if err != nil {
		return fmt.Errorf("update error: Profile does not exist gid: %d, id: %d: %w", profile.GroupId, profile.Id, err)
	}

	profile_json, err := json.MarshalIndent(profile, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(profile_config_path, profile_json, 0666)
	if err != nil {
		return fmt.Errorf("update error: failed to write %s: %w", profile_config_path, err)
	}
	return nil
}

func (db *DB) GetProfile(group_id int, id int) (structs.Profile, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.getProfile(group_id, id)
}

func (db *DB) DeleteProfile(group_id int, id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.deleteProfile(group_id, id)
}

func (db *DB) deleteProfile(group_id int, id int) error {
	profile_config_path := db.GetProfileFilePath(group_id, id)
	err := os.Remove(profile_config_path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Profile removal warning: %s did not exist in the first place", profile_config_path)
		} else {
			return fmt.Errorf("Failed to delete config file %w", err)
		}
	}
	return nil
}

func (db *DB) AddProfile(data structs.DBAddProfileData) (structs.Profile, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.addProfile(data)
}

func (db *DB) getProfile(group_id int, id int) (structs.Profile, error) {
	var profile structs.Profile
	profile_config_path := db.GetProfileFilePath(group_id, id)
	data, err := os.ReadFile(profile_config_path)
	if err != nil {
		return profile, fmt.Errorf("Error reading profile config with gid: %d, id: %d: %w", group_id, id, err)
	}
	if err := json.Unmarshal(data, &profile); err != nil {
		return profile, fmt.Errorf("Error reading profile config with gid: %d, id: %d: %w", group_id, id, err)
	}
	return profile, nil
}

func (db *DB) addProfile(data structs.DBAddProfileData) (structs.Profile, error) {
	oldUmask := syscall.Umask(0)
	defer syscall.Umask(oldUmask)

	var profile_added structs.Profile
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
		GroupId:  data.GroupId,
		Name:     data.Name,
		Protocol: data.Protocol,
		Uri:      data.Uri,
		Host:     data.Host,
		Address:  data.Address,
	}
	profile_json, err := json.MarshalIndent(profile, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(profile_path, profile_json, 0666)
	if err != nil {
		return profile_added, fmt.Errorf("failed to write %s: %w", profile_path, err)
	}

	profile_added = structs.Profile{
		Uri:        data.Uri,
		GroupId:    data.GroupId,
		Id:         profile_id,
		Protocol:   profile.Protocol,
		Name:       profile.Name,
		Host:       profile.Host,
		Address:    profile.Address,
		TestResult: 0,
	}

	return profile_added, nil
}
