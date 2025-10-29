package handler

type Channel struct {
	ID    uint
	Name  string
	URL   string
	M3U8  string
	Proxy bool
	Icon  string
	Epg   string
}

type Config struct {
	BaseURL     string
	Cmd         string
	Args        string
	SecurityKey string
}
