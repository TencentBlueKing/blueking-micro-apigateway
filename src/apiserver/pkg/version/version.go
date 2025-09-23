/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关(BlueKing - Micro APIGateway) available.
 * Copyright (C) 2025 Tencent. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 *     http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 * to the current version of the project delivered to anyone in the future.
 */

// Package version 提供版本信息
package version

import (
	_ "embed"
	"fmt"
	"regexp"
	"runtime"
	"strings"
)

// Version ...
var (
	// Version 版本号
	Version = ""
	// GitCommit CommitID
	GitCommit = ""
	// BuildTime 二进制构建时间
	BuildTime = ""
	// GoVersion Go 版本号
	GoVersion = runtime.Version()
)

//go:embed ChangeLog.md
var changeLogContent []byte

// VersionInfo represents a parsed version release entry
type VersionInfo struct {
	Version string `json:"version"`
	Date    string `json:"date"`
	Content string `json:"content"`
}

// GetVersion 获取版本信息
func GetVersion() string {
	return fmt.Sprintf(
		"\nVersion  : %s\nGitCommit: %s\nBuildTime: %s\nGoVersion: %s\n",
		Version, GitCommit, BuildTime, GoVersion,
	)
}

// GetVersionLog 获取ChangeLog
func GetVersionLog() ([]VersionInfo, error) {
	text := strings.TrimSpace(strings.Replace(string(changeLogContent), "# ChangeLog", "", 1))
	mdFileDatePattern := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	mdFileVersionPattern := regexp.MustCompile(`[vV]\d+\.\d+\.\d+`)
	var versionLogs []VersionInfo
	logs := strings.Split(text, "---")
	for _, log := range logs {
		parts := strings.Split(strings.TrimSpace(log), "\n")
		if len(parts) < 2 {
			continue // Skip logs that don't have enough parts
		}
		dateMatches := mdFileDatePattern.FindStringSubmatch(parts[0])
		versionMatches := mdFileVersionPattern.FindStringSubmatch(parts[1])

		if len(dateMatches) == 0 || len(versionMatches) == 0 {
			continue // Skip if we can't find date or version
		}

		date := dateMatches[0]
		version := versionMatches[0]
		content := strings.Join(parts[2:], "\n") // Combine the rest as content
		versionLogs = append(versionLogs, VersionInfo{
			Version: version,
			Date:    date,
			Content: content,
		})
	}
	return versionLogs, nil
}
