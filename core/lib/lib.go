package lib

import (
	"bushuray-core/db"
	"bushuray-core/structs"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

type profileMetaData struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
}

func AddProfiles(DB *db.DB, data structs.AddProfilesData) structs.ProfilesAdded {
	uris := strings.FieldsSeq(data.Uris)
	var profiles []structs.ProfileAdded
	for uri := range uris {
		log.Println("adding uri:", uri)
		profile_data, err := addProfile(DB, uri, data.GroupId)
		if err == nil {
			profiles = append(profiles, profile_data)
		}
	}
	return structs.ProfilesAdded{
		Profiles: profiles,
	}
}

func addProfile(DB *db.DB, uri string, group_id int) (structs.ProfileAdded, error) {
	v2parserbin := path.Join(getWorkingDir(), "bin", "v2parser")
	v2parser_metadata_cmd := exec.Command(v2parserbin, uri, "--get-metadata")
	var profile_data structs.ProfileAdded
	metadata_output, err := v2parser_metadata_cmd.Output()
	if err != nil {
		return profile_data, fmt.Errorf("getting metadata failed: %w", err)
	}

	var profile_metadata profileMetaData
	if err := json.Unmarshal(metadata_output, &profile_metadata); err != nil {
		return profile_data, fmt.Errorf("err unmarshaling metadata output: %w", err)
	}

	profile_data, err = DB.AddProfile(structs.DBAddProfileData{
		Protocol: profile_metadata.Protocol,
		Name:     profile_metadata.Name,
		Uri:      uri,
		GroupId:  group_id,
	})

	if err != nil {
		return profile_data, fmt.Errorf("err adding profile: %w", err)
	}
	return profile_data, nil

}

func getWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return dir
}
