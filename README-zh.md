# LiveTV

将 Youtube 直播作为 IPTV 源
这个项目分支于[livetv](https://github.com/NatoriMisong/livetv)

感谢 @zjyl1994


## 构建

首先你需要安装Docker

clone 本存储库，执行
`docker build -t livetv .`

构建好容器镜像后，只需要使用以下命令即可在本地的9500端口访问LiveTV!

`docker run -d -p 9500:9000 -v /mnt/data/livetv:/root/data zjyl1994/livetv:1.1`


這將在 9500 連接埠開啓一個使用 `/mnt/data/livetv` 目錄作爲存儲的 LiveTV！ 容器。

PS: 如果不指定外部存儲目錄，LiveTV！重新啓動時將無法讀取之前的設定檔。

## 使用方法

默認的登入密碼是 "password",爲了你的安全請及時修改。

首先你要知道如何在外界訪問到你的主機，如果你使用 VPS 或者獨立伺服器，可以訪問 `http://你的主機ip:9500`，你應該可以看到以下畫面：

![index_page](pic/index-zh.png)

首先你需要在設定區域點擊“自動填充”，設定正確的URL。然後點擊“儲存設定”。

然後就可以添加頻道，頻道添加成功后就能M3U8檔案列的地址進行播放了。

當你使用Kodi之類的播放器，可以考慮使用第一行的M3U檔案URL進行播放，會自動生成包含所有頻道信息的播放列表。

Youtube-dl的文檔可以在這裏找到 => [https://github.com/ytdl-org/youtube-dl](https://github.com/ytdl-org/youtube-dl)



当你不能直接链接Youtube时，你需要首先有一个飞机。具体怎么弄自己查，此处假定你的飞机地址为 socks5://192.168.1.1:10808。

运行 docker run -p9500:9000 -eHTTP_PROXY=socks5://192.168.1.1:10808 -eHTTPS_PROXY=socks5://192.168.1.1:10808 -v/mnt/d/workspace/livetv/data:/root/data zjyl1994/livetv:1.0