package ui

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

const (
	cnbReleasesAPI = "https://cnb.cool/dmxapi/opencode_dmxapi/-/releases"
	downloadURL    = "https://cnb.cool/dmxapi/opencode_dmxapi/-/releases"
)

// UpdateResult 存储版本检查结果
type UpdateResult struct {
	HasUpdate     bool
	LatestVersion string
	DownloadURL   string
}

// CheckForUpdateAsync 异步检查 CNB 最新版本，通过 channel 返回结果
// 失败时发送 UpdateResult{HasUpdate: false}
func CheckForUpdateAsync() <-chan UpdateResult {
	ch := make(chan UpdateResult, 1)
	go func() {
		ch <- checkUpdate()
	}()
	return ch
}

func checkUpdate() UpdateResult {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", cnbReleasesAPI, nil)
	if err != nil {
		return UpdateResult{}
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return UpdateResult{}
	}
	defer resp.Body.Close()

	var releases []struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return UpdateResult{}
	}

	if len(releases) == 0 {
		return UpdateResult{}
	}

	latestTag := strings.TrimPrefix(releases[0].TagName, "v")
	if latestTag == "" || latestTag == Version {
		return UpdateResult{}
	}

	return UpdateResult{
		HasUpdate:     true,
		LatestVersion: latestTag,
		DownloadURL:   downloadURL,
	}
}
