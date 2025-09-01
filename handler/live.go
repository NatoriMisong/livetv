package handler

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/NatoriMisong/livetv/global"
	"github.com/NatoriMisong/livetv/service"
	"github.com/NatoriMisong/livetv/util"
)

func M3UHandler(c *gin.Context) {
	// 验证security_key
	securityKey := c.Query("k")
	actualKey, err := service.GetConfig("security_key")
	if err != nil || securityKey != actualKey {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	
	content, err := service.M3UGenerate()
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/vnd.apple.mpegurl", []byte(content))
}

func LiveHandler(c *gin.Context) {
	// 验证security_key
	securityKey := c.Query("k")
	actualKey, err := service.GetConfig("security_key")
	if err != nil || securityKey != actualKey {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	
	var m3u8Body string
	channelCacheKey := c.Query("c")
	iBody, found := global.M3U8Cache.Get(channelCacheKey)
	if found {
		m3u8Body = iBody.(string)
	} else {
		channelNumber := util.String2Uint(c.Query("c"))
		if channelNumber == 0 {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		channelInfo, err := service.GetChannel(channelNumber)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				c.AbortWithStatus(http.StatusNotFound)
			} else {
				log.Println(err)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
			return
		}
		baseUrl, err := service.GetConfig("base_url")
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		liveM3U8, err := service.GetYoutubeLiveM3U8(channelInfo.URL)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		client := http.Client{Timeout: global.HttpClientTimeout}
		resp, err := client.Get(liveM3U8)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		bodyString := string(bodyBytes)
		if channelInfo.Proxy {
			// 记录处理前的URL信息
			log.Printf("Processing M3U8 for channel %d with proxy, baseUrl: %s", channelNumber, baseUrl)
			m3u8Body = service.M3U8Process(bodyString, baseUrl+"/live.ts?", securityKey)
			// 记录处理后的前几行内容用于调试
			lines := strings.SplitN(m3u8Body, "\n", 5)
			log.Printf("Processed M3U8 first lines: %s", strings.Join(lines, "\n"))
		} else {
			m3u8Body = bodyString
		}
		global.M3U8Cache.Set(channelCacheKey, m3u8Body, 3*time.Second)
	}
	// 设置正确的MIME类型和Cache-Control头
	c.Header("Content-Type", "application/vnd.apple.mpegurl")
	c.Header("Cache-Control", "no-cache")
	c.String(http.StatusOK, m3u8Body)
}

func TsProxyHandler(c *gin.Context) {
	// 验证security_key
	securityKey := c.Query("k")
	actualKey, err := service.GetConfig("security_key")
	if err != nil || securityKey != actualKey {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	
	// 获取和解压远程URL
	zipedRemoteURL := c.Query("url")
	remoteURL, err := util.DecompressString(zipedRemoteURL)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if remoteURL == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	client := http.Client{Timeout: global.HttpClientTimeout}
	resp, err := client.Get(remoteURL)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	c.DataFromReader(http.StatusOK, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}

func CacheHandler(c *gin.Context) {
	var sb strings.Builder
	global.URLCache.Range(func(k, v interface{}) bool {
		sb.WriteString(k.(string))
		sb.WriteString(" => ")
		sb.WriteString(v.(string))
		sb.WriteString("\n")
		return true
	})
	c.Data(http.StatusOK, "text/plain", []byte(sb.String()))
}
