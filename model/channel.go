package model

type Channel struct {
	ID    uint `gorm:"primary_key"`
	Name  string
	URL   string
	Proxy bool
	Icon  string
	Epg   string
}
