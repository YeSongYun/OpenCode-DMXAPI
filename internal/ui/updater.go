package ui

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
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
	req.Header.Set("User-Agent", "opencode-dmxapi/"+Version)

	resp, err := client.Do(req)
	if err != nil {
		return UpdateResult{}
	}
	defer resp.Body.Close()

	var releases []struct {
		TagName string `json:"tag_name"`
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return UpdateResult{}
	}
	if err := json.Unmarshal(body, &releases); err != nil {
		return UpdateResult{}
	}

	if len(releases) == 0 {
		return UpdateResult{}
	}

	latestTag := strings.TrimPrefix(releases[0].TagName, "v")
	if latestTag == "" || !isNewerVersion(latestTag, Version) {
		return UpdateResult{}
	}

	return UpdateResult{
		HasUpdate:     true,
		LatestVersion: latestTag,
		DownloadURL:   downloadURL,
	}
}

// isNewerVersion 判断 latest 是否比 current 更新。
// 解析失败时回退到字符串不等比较；current == "dev" 时永不提示（避免本地开发构建噪音）。
func isNewerVersion(latest, current string) bool {
	if current == "dev" {
		return false
	}
	lp, lok := parseSemver(latest)
	cp, cok := parseSemver(current)
	if !lok || !cok {
		return latest != current
	}
	for i := 0; i < 3; i++ {
		if lp[i] != cp[i] {
			return lp[i] > cp[i]
		}
	}
	return false
}

// parseSemver 将 "x.y.z" 解析为 [3]int。
// 容忍预发布/构建元数据后缀（如 "2.1.0-beta.1"、"2.1.0+sha"）：每段只取前导数字。
// 任一段缺少前导数字时返回 ok=false。
func parseSemver(s string) ([3]int, bool) {
	var out [3]int
	parts := strings.SplitN(s, ".", 3)
	if len(parts) < 1 {
		return out, false
	}
	for i := 0; i < 3 && i < len(parts); i++ {
		seg := strings.TrimSpace(parts[i])
		end := strings.IndexFunc(seg, func(r rune) bool { return r < '0' || r > '9' })
		if end == 0 {
			return out, false
		}
		if end > 0 {
			seg = seg[:end]
		}
		n, err := strconv.Atoi(seg)
		if err != nil {
			return out, false
		}
		out[i] = n
	}
	return out, true
}
