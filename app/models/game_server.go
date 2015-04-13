package models

import (
	"fmt"
	"github.com/revel/revel"
	"gopkg.in/yaml.v2"
	"os"
)

type GameServerConfig struct {
	Ip     string `ip`
	Port   int    `port`
	ZoneId int    `zoneId`
	Domain string `domain`
	Name   string `name`
}

// 每日刷新玩家基本数据url
func (g *GameServerConfig) UserRankUrl(id int) string {
	return fmt.Sprintf("http://%s:%d/%s/admin/wwa/gr?type=%d", g.Ip, g.Port, g.Domain, id)
}

// 获取玩家战斗Group数据
func (g *GameServerConfig) GroupDataUrl(t int) string {
	return fmt.Sprintf("http://%s:%d/%s/admin/wwa/gd?id=%d", g.Ip, g.Port, g.Domain, t)
}

// 获取玩家战斗Group数据
func (g *GameServerConfig) GroupInfoUrl(t int) string {
	return fmt.Sprintf("http://%s:%d/%s/admin/wwa/gi?id=%d", g.Ip, g.Port, g.Domain, t)
}

func (g GameServerConfig) String() string {
	return fmt.Sprintf("ip=%s,port=%d,zoneId=%d\n", g.Ip, g.Port, g.ZoneId)
}

var (
	GameServerList []GameServerConfig
)

func FindGameServer(zoneId int) *GameServerConfig {
	for _, g := range GameServerList {
		if g.ZoneId == zoneId {
			return &g
		}
	}
	return nil
}

func InitGameServerConfig() {
	root, ok := revel.Config.String("root")
	if !ok {
		panic("app.conf 中没有配置root，项目绝对路径")
	}
	filepath := fmt.Sprintf("%s/conf/game_server.yml", root)
	file, err := os.Open(filepath)
	if err != nil {
		panic(fmt.Sprintf("没有找到配置文件%s", filepath))
	}
	fi, _ := file.Stat()
	size := fi.Size()
	data := make([]byte, size)
	file.Read(data)
	fmt.Printf(" size = %d, yml : %s\n", size, data)
	yaml.Unmarshal(data, &GameServerList)
	fmt.Println("gamserverList", GameServerList)
}
