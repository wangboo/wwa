package models

import (
	"fmt"
	"github.com/revel/revel"
	"gopkg.in/yaml.v2"
	"log"
	"net/url"
	"os"
)

type GameServerConfig struct {
	Ip     string `yaml:"ip"`
	Port   int    `yaml:"port"`
	ZoneId int    `yaml:"zoneId"`
	Domain string `yaml:"domain"`
	Name   string `yaml:"name"`
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

// 获取玩家战斗Group数据
func (g *GameServerConfig) MuInfoUrl(t int) string {
	return fmt.Sprintf("http://%s:%d/%s/admin/wwa/mu?id=%d", g.Ip, g.Port, g.Domain, t)
}

// 获取玩家更换服务器
func (g *GameServerConfig) ChangeServerUrl(name string) string {
	return fmt.Sprintf("http://%s:%d/%s/admin/master/changeServer?token=%s", g.Ip, g.Port, g.Domain, name)
}

func (g *GameServerConfig) UserLevelAndVipUrl(username string) string {
	return fmt.Sprintf("http://%s:%d/%s/admin/inv/levelAndVip?username=%s", g.Ip, g.Port, g.Domain, username)
}

func (g *GameServerConfig) UserInvInfoUrl(username string) string {
	return fmt.Sprintf("http://%s:%d/%s/admin/inv/info?username=%s", g.Ip, g.Port, g.Domain, username)
}

func (g GameServerConfig) String() string {
	return fmt.Sprintf("ip=%s,port=%d,zoneId=%d\n", g.Ip, g.Port, g.ZoneId)
}

// 通知游戏服务器发送邮件
func (g *GameServerConfig) MailUrl(userId int, mail string) string {
	return fmt.Sprintf("http://%s:%d/%s/admin/mail?u=%d&mail=%s", g.Ip, g.Port, g.Domain, userId, mail)
}

// 日终奖励
func (g *GameServerConfig) DayEndWwaRewardUrl(str string, Type int) string {
	return fmt.Sprintf("http://%s:%d/%s/admin/wwa/dayEndReward?t=%d&base64=%s", g.Ip, g.Port, g.Domain, Type, str)
}

// 通用奖励邮件
func (g *GameServerConfig) CommonRewardMail(recv int, msg, reward string) string {
	// crm = CommonRewardMail
	msg = url.QueryEscape(msg)
	reward = url.QueryEscape(reward)
	return fmt.Sprintf("http://%s:%d/%s/admin/wwa/crm?msg=%s&reward=%s&recvId=%d", g.Ip, g.Port, g.Domain, msg, reward, recv)
}

func (g *GameServerConfig) Payment() string {
	return fmt.Sprintf("http://%s:%d/%s/admin/charge", g.Ip, g.Port, g.Domain)
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
	defer file.Close()
	fi, _ := file.Stat()
	size := fi.Size()
	data := make([]byte, size)
	file.Read(data)
	fmt.Printf(" size = %d, yml : %s\n", size, data)
	err = yaml.Unmarshal(data, &GameServerList)
	if err != nil {
		log.Panicf("load game_server.yml err %s\n", err.Error())
		return
	}
	for _, gs := range GameServerList {
		log.Printf("gs: ip=%s,port=%d,zoneId=%d,domain=%s,name=%s\n", gs.Ip, gs.Port, gs.ZoneId, gs.Domain, gs.Name)
	}
}
