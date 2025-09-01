package handler

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/NatoriMisong/livetv/model"
	"github.com/NatoriMisong/livetv/service"
	"github.com/NatoriMisong/livetv/util"

	"golang.org/x/text/language"
)

var langMatcher = language.NewMatcher([]language.Tag{
	language.English,
	language.Chinese,
})

func IndexHandler(c *gin.Context) {
	if sessions.Default(c).Get("logined") != true {
		c.Redirect(http.StatusFound, "/login")
	}
	acceptLang := c.Request.Header.Get("Accept-Language")
	langTag, _ := language.MatchStrings(langMatcher, acceptLang)

	baseUrl, err := service.GetConfig("base_url")
	if err != nil {
		log.Println(err.Error())
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrMsg": err.Error(),
		})
		return
	}
	channelModels, err := service.GetAllChannel()
	if err != nil {
		log.Println(err.Error())
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrMsg": err.Error(),
		})
		return
	}
	var m3uName string
	if langTag == language.Chinese {
		m3uName = "M3U 頻道列表"
	} else {
		m3uName = "M3U File"
	}
	channels := make([]Channel, len(channelModels)+1)
	channels[0] = Channel{
		ID:   0,
		Name: m3uName,
		M3U8: baseUrl + "/lives.m3u",
	}
	for i, v := range channelModels {
		channels[i+1] = Channel{
			ID:    v.ID,
			Name:  v.Name,
			URL:   v.URL,
			M3U8:  baseUrl + "/live.m3u8?c=" + strconv.Itoa(int(v.ID)),
			Proxy: v.Proxy,
		}
	}
	conf, err := loadConfig()
	if err != nil {
		log.Println(err.Error())
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrMsg": err.Error(),
		})
		return
	}

	var templateFilename string
	if langTag == language.Chinese {
		templateFilename = "index-zh.html"
	} else {
		templateFilename = "index.html"
	}
	c.HTML(http.StatusOK, templateFilename, gin.H{
		"Channels": channels,
		"Configs":  conf,
	})
}

func loadConfig() (Config, error) {
	var conf Config
	if cmd, err := service.GetConfig("ytdl_cmd"); err != nil {
		return conf, err
	} else {
		conf.Cmd = cmd
	}
	if args, err := service.GetConfig("ytdl_args"); err != nil {
		return conf, err
	} else {
		conf.Args = args
	}
	if burl, err := service.GetConfig("base_url"); err != nil {
		return conf, err
	} else {
		conf.BaseURL = burl
	}
	return conf, nil
}

func NewChannelHandler(c *gin.Context) {
	if sessions.Default(c).Get("logined") != true {
		c.Redirect(http.StatusFound, "/login")
	}
	chName := c.PostForm("name")
	chURL := c.PostForm("url")
	if chName == "" || chURL == "" {
		c.Redirect(http.StatusFound, "/")
		return
	}
	chProxy := c.PostForm("proxy") != ""
	mch := model.Channel{
		Name:  chName,
		URL:   chURL,
		Proxy: chProxy,
	}
	err := service.SaveChannel(mch)
	if err != nil {
		log.Println(err.Error())
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrMsg": err.Error(),
		})
		return
	}
	c.Redirect(http.StatusFound, "/")
}

func DeleteChannelHandler(c *gin.Context) {
	if sessions.Default(c).Get("logined") != true {
		c.Redirect(http.StatusFound, "/login")
	}
	chID := util.String2Uint(c.Query("id"))
	if chID == 0 {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrMsg": "empty id",
		})
		return
	}
	err := service.DeleteChannel(chID)
	if err != nil {
		log.Println(err.Error())
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrMsg": err.Error(),
		})
		return
	}
	c.Redirect(http.StatusFound, "/")
}

func UpdateChannelHandler(c *gin.Context) {
	if sessions.Default(c).Get("logined") != true {
		c.Redirect(http.StatusFound, "/login")
	}
	chID := util.String2Uint(c.PostForm("id"))
	if chID == 0 {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrMsg": "empty id",
		})
		return
	}
	chName := c.PostForm("name")
	chURL := c.PostForm("url")
	chProxy := c.PostForm("proxy") != ""
	chIndexStr := c.PostForm("index")
	if chName == "" || chURL == "" {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrMsg": "name or url cannot be empty",
		})
		return
	}
	// 获取现有频道
	channel, err := service.GetChannel(chID)
	if err != nil {
		log.Println(err.Error())
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrMsg": err.Error(),
		})
		return
	}
	// 更新频道信息
	channel.Name = chName
	channel.URL = chURL
	channel.Proxy = chProxy
	
	// 处理序号更新
	if chIndexStr != "" {
		chIndex, err := strconv.ParseUint(chIndexStr, 10, 64)
		if err == nil {
			// 如果序号有效且不为0，则进行序号更新
			if chIndex > 0 {
				// 首先获取当前所有频道
				allChannels, err := service.GetAllChannel()
				if err != nil {
					log.Println(err.Error())
					c.HTML(http.StatusInternalServerError, "error.html", gin.H{
						"ErrMsg": err.Error(),
					})
					return
				}
				
				// 更新序号逻辑
				// 如果新序号小于或等于当前所有频道数量，调整受影响的频道序号
				if uint64(len(allChannels)) >= chIndex {
					// 查找新序号位置的频道
					for i := range allChannels {
						if allChannels[i].ID == channel.ID {
							// 跳过当前正在编辑的频道
							continue
						}
						// 调整其他频道的序号
						if (channel.ID < allChannels[i].ID && uint64(i+1) >= chIndex) || (channel.ID > allChannels[i].ID && uint64(i+1) >= chIndex) {
							if uint64(i+1) >= chIndex && allChannels[i].ID != channel.ID {
								if channel.ID > allChannels[i].ID {
									// 向前移动
									allChannels[i].ID--
								} else {
									// 向后移动
									allChannels[i].ID++
								}
								// 保存更新后的频道
								err := service.SaveChannel(allChannels[i])
								if err != nil {
									log.Println(err.Error())
									c.HTML(http.StatusInternalServerError, "error.html", gin.H{
										"ErrMsg": err.Error(),
									})
									return
								}
							}
					}
				}
				// 设置当前频道的新序号
				channel.ID = uint(chIndex)
			}
		}
	}
	
	err = service.SaveChannel(channel)
	if err != nil {
		log.Println(err.Error())
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrMsg": err.Error(),
		})
		return
	}
	c.Redirect(http.StatusFound, "/")
}

func UpdateConfigHandler(c *gin.Context) {
	if sessions.Default(c).Get("logined") != true {
		c.Redirect(http.StatusFound, "/login")
	}
	ytdlCmd := c.PostForm("cmd")
	ytdlArgs := c.PostForm("args")
	baseUrl := strings.TrimSuffix(c.PostForm("baseurl"), "/")
	if len(ytdlCmd) > 0 {
		err := service.SetConfig("ytdl_cmd", ytdlCmd)
		if err != nil {
			log.Println(err.Error())
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"ErrMsg": err.Error(),
			})
			return
		}
	}
	if len(ytdlArgs) > 0 {
		err := service.SetConfig("ytdl_args", ytdlArgs)
		if err != nil {
			log.Println(err.Error())
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"ErrMsg": err.Error(),
			})
			return
		}
	}
	if len(baseUrl) > 0 {
		err := service.SetConfig("base_url", baseUrl)
		if err != nil {
			log.Println(err.Error())
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"ErrMsg": err.Error(),
			})
			return
		}
	}
	c.Redirect(http.StatusFound, "/")
}

func LogHandler(c *gin.Context) {
	if sessions.Default(c).Get("logined") != true {
		c.Redirect(http.StatusFound, "/login")
	}
	c.File(os.Getenv("LIVETV_DATADIR") + "/livetv.log")
}

func LoginViewHandler(c *gin.Context) {
	session := sessions.Default(c)
	crsfToken := util.RandString(10)
	session.Set("crsfToken", crsfToken)
	err := session.Save()
	if err != nil {
		log.Println(err.Error())
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrMsg": err.Error(),
		})
		return
	}
	c.HTML(http.StatusOK, "login.html", gin.H{
		"Crsf": crsfToken,
	})
}

func LoginActionHandler(c *gin.Context) {
	session := sessions.Default(c)
	crsfToken := c.PostForm("crsf")
	if crsfToken != session.Get("crsfToken") {
		c.HTML(http.StatusOK, "error.html", gin.H{
			"ErrMsg": "Password error!",
		})
		return
	}
	pass := c.PostForm("password")
	cfgPass, err := service.GetConfig("password")
	if err != nil {
		log.Println(err.Error())
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrMsg": err.Error(),
		})
		return
	}
	if pass == cfgPass {
		session.Set("logined", true)
		err = session.Save()
		if err != nil {
			log.Println(err.Error())
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"ErrMsg": err.Error(),
			})
			return
		}
		c.Redirect(http.StatusFound, "/")
	} else {
		c.HTML(http.StatusOK, "error.html", gin.H{
			"ErrMsg": "Password error!",
		})
	}
}

func LogoutHandler(c *gin.Context) {
	if sessions.Default(c).Get("logined") != true {
		c.Redirect(http.StatusFound, "/login")
	}
	session := sessions.Default(c)
	session.Delete("logined")
	err := session.Save()
	if err != nil {
		log.Println(err.Error())
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrMsg": err.Error(),
		})
		return
	}
	c.Redirect(http.StatusFound, "/login")
}

func ChangePasswordHandler(c *gin.Context) {
	if sessions.Default(c).Get("logined") != true {
		c.Redirect(http.StatusFound, "/login")
	}
	pass := c.PostForm("password")
	pass2 := c.PostForm("password2")
	if pass != pass2 {
		c.HTML(http.StatusOK, "error.html", gin.H{
			"ErrMsg": "Password mismatch!",
		})
	}
	err := service.SetConfig("password", pass)
	if err != nil {
		log.Println(err.Error())
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrMsg": err.Error(),
		})
		return
	}
	LogoutHandler(c)
}
