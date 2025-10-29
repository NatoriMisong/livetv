package service

import (
	"bufio"
	"strings"

	"github.com/NatoriMisong/livetv/util"
)

func M3U8Process(data string, prefixURL string, securityKey string) string {
	var sb strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		l := scanner.Text()
		if strings.HasPrefix(l, "#") {
			sb.WriteString(l)
		} else {
			sb.WriteString(prefixURL)
			sb.WriteString("k=")
			sb.WriteString(securityKey)
			sb.WriteString("&url=")
			sb.WriteString(util.CompressString(l))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
