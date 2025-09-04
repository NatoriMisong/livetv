package service

import (
	"log"
	"strconv"
	"strings"
)

func M3UGenerate() (string, error) {
	baseUrl, err := GetConfig("base_url")
	if err != nil {
		log.Println(err)
		return "", err
	}
	// 获取安全密钥
	securityKey, err := GetConfig("security_key")
	if err != nil {
		log.Println(err)
		return "", err
	}
	channels, err := GetAllChannel()
	if err != nil {
		log.Println(err)
		return "", err
	}
	var m3u strings.Builder
	m3u.WriteString("#EXTM3U\n")
	for _, v := range channels {
		m3u.WriteString("#EXTINF:-1")
		// 添加图标信息
		if v.Icon != "" {
			m3u.WriteString(" tvg-logo=\"" + v.Icon + "\"")
		}
		// 添加EPG信息
		if v.Epg != "" {
			m3u.WriteString(" tvg-url=\"" + v.Epg + "\"")
		}
		m3u.WriteString(",")
		m3u.WriteString(v.Name)
		m3u.WriteString("\n")
		m3u.WriteString(baseUrl)
		m3u.WriteString("/live.m3u8?c=")
		m3u.WriteString(strconv.Itoa(int(v.ID)))
		m3u.WriteString("&k=")
		m3u.WriteString(securityKey)
		m3u.WriteString("\n")
	}
	return m3u.String(), nil
}
