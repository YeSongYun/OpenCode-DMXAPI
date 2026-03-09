package ui

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

const (
	githubReleasesAPI = "https://api.github.com/repos/YeSongYun/OpenCode-DMXAPI/releases/latest"
	downloadURL       = "https://cnb.cool/dmxapi/opencode_dmxapi/-/releases"
)

// UpdateResult 存储版本检查结果
type UpdateResult struct {
	HasUpdate     bool
	LatestVersion string
	DownloadURL   string
}

// CheckForUpdateAsync 异步检查 GitHub 最新版本，通过 channel 返回结果
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
	resp, err := client.Get(githubReleasesAPI)
	if err != nil {
		return UpdateResult{}
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return UpdateResult{}
	}

	latestTag := strings.TrimPrefix(release.TagName, "v")
	if latestTag == "" || latestTag == Version {
		return UpdateResult{}
	}

	return UpdateResult{
		HasUpdate:     true,
		LatestVersion: latestTag,
		DownloadURL:   downloadURL,
	}
}
