package cmd

import (
	"bushuray-core/db"
	"bushuray-core/structs"
	"encoding/json"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
)

type profileMetaData struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
}

type Cmd struct {
	Conn net.Conn
	DB   *db.DB
}

func (cmd *Cmd) AddProfile(data structs.AddProfileData) {
	v2parserbin := path.Join(getWorkingDir(), "bin", "v2parser")
	v2parser_cmd := exec.Command(v2parserbin, data.Uri, "--get-metadata")
	metadata_output, err := v2parser_cmd.Output()
	if err != nil {
		log.Println("getting metadata failed:", err)
		return
	}

	var profile_metadata profileMetaData
	if err := json.Unmarshal(metadata_output, &profile_metadata); err != nil {
		log.Println("err unmarshaling metadata output")
		return
	}

	profile_data, err := cmd.DB.AddProfile(structs.DBAddProfileData{
		Protocol: profile_metadata.Protocol,
		Name:     profile_metadata.Name,
		Uri:      data.Uri,
		GroupId:  data.GroupId,
	})

	if err != nil {
		log.Println("err adding profile:", err)
	} else {
		log.Println(profile_data)
	}

	// var raw json.RawMessage
	// if err := json.Unmarshal(output, &raw); err != nil {
	// 	log.Println("err unmarshaling json")
	// 	return
	// }
}

func getWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return dir
}
