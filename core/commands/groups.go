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

	subscription_content, err := get(group.SubscriptionUrl, cmd.AppConfig.SubscriptionUserAgent)
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

func get(url string, userAgent string) (string, error) {
	req, err := newSubscriptionRequest(url, userAgent)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
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

func newSubscriptionRequest(url string, userAgent string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if userAgent == "" {
		req.Header["User-Agent"] = nil
	} else {
		req.Header.Set("User-Agent", userAgent)
	}

	return req, nil
}
