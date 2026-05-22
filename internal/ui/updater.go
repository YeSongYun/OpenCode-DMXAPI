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

	// 从 releases 中过滤掉预发布版本（tag 含 "-"），用 parseSemver 选最大稳定版。
	// 不依赖 releases[0]，避免 API 返回顺序变化或维护者补发旧版导致误判。
	var (
		latestTag string
		latestVer [3]int
		found     bool
	)
	for _, r := range releases {
		tag := strings.TrimPrefix(r.TagName, "v")
		if tag == "" || strings.Contains(tag, "-") {
			continue
		}
		v, ok := parseSemver(tag)
		if !ok {
			continue
		}
		if !found || compareSemver(v, latestVer) > 0 {
			latestTag = tag
			latestVer = v
			found = true
		}
	}
	if !found {
		return UpdateResult{}
	}

	if !isNewerVersion(latestTag, Version) {
		return UpdateResult{}
	}

	return UpdateResult{
		HasUpdate:     true,
		LatestVersion: latestTag,
		DownloadURL:   downloadURL,
	}
}

// compareSemver 比较两个 [3]int 版本号，a > b 返回 1，a < b 返回 -1，相等返回 0
func compareSemver(a, b [3]int) int {
	for i := 0; i < 3; i++ {
		if a[i] != b[i] {
			if a[i] > b[i] {
				return 1
			}
			return -1
		}
	}
	return 0
}

// isNewerVersion 判断 latest 是否比 current 更新。
// 解析失败时回退到字符串不等比较；current == "dev" 时永不提示（避免本地开发构建噪音）；
// latest 含 "-"（预发布）时永不提示稳定用户。
func isNewerVersion(latest, current string) bool {
	if current == "dev" {
		return false
	}
	if strings.Contains(latest, "-") {
		return false
	}
	lp, lok := parseSemver(latest)
	cp, cok := parseSemver(current)
	if !lok || !cok {
		return latest != current
	}
	return compareSemver(lp, cp) > 0
}

// parseSemver 将 "x.y.z" 解析为 [3]int。
// 容忍预发布/构建元数据后缀（如 "2.1.0-beta.1"、"2.1.0+sha"）：每段只取前导数字。
// 严格要求 3 段，"2.1" / "v2" 这类短版本号返回 ok=false，避免静默补 0 绕过过滤。
func parseSemver(s string) ([3]int, bool) {
	var out [3]int
	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return out, false
	}
	for i := 0; i < 3; i++ {
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
