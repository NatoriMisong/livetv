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
)

func IndexHandler(c *gin.Context) {
	if sessions.Default(c).Get("logined") != true {
		c.Redirect(http.StatusFound, "/login")
	}

	baseUrl, err := service.GetConfig("base_url")
	if err != nil {
		log.Println(err.Error())
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrMsg": err.Error(),
		})
		return
	}
	// 获取安全密钥
	securityKey, err := service.GetConfig("security_key")
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
	m3uName := "M3U File"
	channels := make([]Channel, len(channelModels)+1)
	channels[0] = Channel{
		ID:   0,
		Name: m3uName,
		M3U8: baseUrl + "/lives.m3u?k=" + securityKey,
	}
	for i, v := range channelModels {
			channels[i+1] = Channel{
				ID:    v.ID,
				Name:  v.Name,
				URL:   v.URL,
				M3U8:  baseUrl + "/live.m3u8?c=" + strconv.Itoa(int(v.ID)) + "&k=" + securityKey,
				Proxy: v.Proxy,
				Icon:  v.Icon,
				Epg:   v.Epg,
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

	// 只使用英文模板
	c.HTML(http.StatusOK, "index.html", gin.H{
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
	// 加载security_key配置，如果不存在则生成一个随机12字符值
	securityKey, err := service.GetConfig("security_key")
	if err != nil {
		// 检查是否是配置不存在的错误
		if strings.Contains(err.Error(), "not found") {
			// 生成12字符随机值
			securityKey = util.RandString(12)
			service.SetConfig("security_key", securityKey)
		} else {
			return conf, err
		}
	}
	conf.SecurityKey = securityKey
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
	chIcon := c.PostForm("icon")
	chEpg := c.PostForm("epg")
	mch := model.Channel{
		Name:  chName,
		URL:   chURL,
		Proxy: chProxy,
		Icon:  chIcon,
		Epg:   chEpg,
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
	channel.Icon = c.PostForm("icon")
	channel.Epg = c.PostForm("epg")
	
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
				
				// 获取频道的当前位置
				var currentPos int
				for i, ch := range allChannels {
					if ch.ID == channel.ID {
						currentPos = i + 1 // 位置从1开始计数
						break
					}
				}
				
				// 如果新序号与当前位置不同，才进行调整
				if uint64(currentPos) != chIndex {
					// 如果新序号大于频道总数，则将其设置为总数+1
					if chIndex > uint64(len(allChannels)) {
						chIndex = uint64(len(allChannels)) + 1
					}
					
					// 调整其他频道的序号
					if uint64(currentPos) < chIndex {
						// 频道向后移动，需要将中间的频道序号减1
						for i := range allChannels {
							if allChannels[i].ID == channel.ID {
								continue // 跳过当前正在编辑的频道
							}
							// 调整在当前位置和新位置之间的频道
							if uint64(i+1) > uint64(currentPos) && uint64(i+1) <= chIndex {
								allChannels[i].ID--
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
					} else if uint64(currentPos) > chIndex {
						// 频道向前移动，需要将中间的频道序号加1
						for i := range allChannels {
							if allChannels[i].ID == channel.ID {
								continue // 跳过当前正在编辑的频道
							}
							// 调整在新位置和当前位置之间的频道
							if uint64(i+1) >= chIndex && uint64(i+1) < uint64(currentPos) {
								allChannels[i].ID++
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
	// 处理security_key配置
	securityKey := c.PostForm("security_key")
	if len(securityKey) > 0 {
		err := service.SetConfig("security_key", securityKey)
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
