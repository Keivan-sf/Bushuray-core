package structs

import "encoding/json"

// database types
type DBAddProfileData struct {
	Uri      string `json:"uri"`
	GroupId  int    `json:"group_id"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
}

// commands and responses
type DeleteProfilesData struct {
	Profiles []ProfileID `json:"profiles"`
}

type ProfilesDeleted struct {
	DeletedProfiles []ProfileID `json:"deleted-profiles"`
}

type AddProfilesData struct {
	Uris    string `json:"uris"`
	GroupId int    `json:"group_id"`
}

type ProfilesAdded struct {
	Profiles []ProfileAdded `json:"profiles"`
}

type ProfileAdded struct {
	Id       int    `json:"id"`
	GroupId  int    `json:"group_id"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Uri      string `json:"uri"`
}

// general types
type Group struct {
	Id              int    `json:"id"`
	SubscriptionUrl string `json:"subscription_url"`
	Name            string `json:"name"`
	LastId          int    `json:"last_id"`
}

type Profile struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Uri      string `json:"uri"`
}

type ProfileID struct {
	Id      int `json:"id"`
	GroupId int `json:"group_id"`
}


type TCPMessage struct {
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}


type Message[T any] struct {
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
