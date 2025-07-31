package cmd

import (
	"bushuray-core/lib"
	"bushuray-core/structs"
	"io"
	"log"
	"net/http"
)

func (cmd *Cmd) UpdateSubscription(data structs.UpdateSubscriptionData) {
	group, err := cmd.DB.LoadGroupConfig(data.GroupId)
	if err != nil {
		return
	}

	subscription_content, err := get(group.SubscriptionUrl)
	if err != nil {
		return
	}
	db_profiles := lib.GetDBAddProfileDatasFromStr(subscription_content, data.GroupId)
	profiles, err := cmd.DB.UpdateGroupAndProfiles(data.GroupId, db_profiles)
	if err != nil {
		log.Println(err)
		return
	}
	cmd.send("subscription-updated", structs.SubscriptionUpdated{Profiles: profiles})

}

func get(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}
