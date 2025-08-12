package cmd

import (
	"bushuray-core/lib"
	proxy "bushuray-core/lib/proxy/mainproxy"
	"bushuray-core/structs"
	"io"
	"log"
	"net/http"
	"sync"
)

var subscription_mutex = sync.Mutex{}

func (cmd *Cmd) UpdateSubscription(data structs.UpdateSubscriptionData, proxy_manager *proxy.ProxyManager) {
	if subscription_mutex.TryLock() == false {
		cmd.warn("command-ignored", "update-subscription")
		return
	}
	defer subscription_mutex.Unlock()

	group, err := cmd.DB.LoadGroupConfig(data.GroupId)
	if err != nil {
		cmd.warn("update-subscription-failed", "Failed to load group config for updating subscription")
		return
	}

	subscription_content, err := get(group.SubscriptionUrl)
	if err != nil {
		cmd.warn("update-subscription-failed", "Failed to get subscription content")
		return
	}

	db_profiles := lib.GetDBAddProfileDatasFromStr(subscription_content, data.GroupId)
	status := proxy_manager.GetStatus()
	keep_profile_id := 0
	if status.Connection == "connected" && status.Profile.GroupId == data.GroupId {
		keep_profile_id = status.Profile.Id
	}

	profiles, err := cmd.DB.UpdateGroupAndProfiles(data.GroupId, db_profiles, keep_profile_id)
	if err != nil {
		log.Println(err)
		cmd.warn("update-subscription-failed", "Failed to add new subscription content to database")
		return
	}
	cmd.send("subscription-updated", structs.SubscriptionUpdated{GroupId: data.GroupId, Profiles: profiles})

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
