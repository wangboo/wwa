package models

import (
	"encoding/base64"
	"fmt"
	"github.com/revel/revel"
	"gopkg.in/yaml.v2"
	"log"
	"net/url"
	"os"
	"time"
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
	msgBase64 := base64.StdEncoding.EncodeToString([]byte(msg))
	msgBase64 = url.QueryEscape(msgBase64)
	reward = url.QueryEscape(reward)
	url := fmt.Sprintf("http://%s:%d/%s/admin/wwa/crm?msg=%s&reward=%s&recvId=%d", g.Ip, g.Port, g.Domain, msgBase64, reward, recv)
	fmt.Println("url = ", url)
	return url
}

// 电视通知消息
func (g *GameServerConfig) NoticeUrl(msg string) string {
	msgBase64 := base64.StdEncoding.EncodeToString([]byte(msg))
	msgBase64 = url.QueryEscape(msgBase64)
	url := fmt.Sprintf("http://%s:%d/%s/admin/www/notice?msg=", g.Ip, g.Port, g.Domain) + msgBase64
	return url
}

func (g *GameServerConfig) NoticeAdvanceUrl(name, msg string) string {
	msgBase64 := base64.StdEncoding.EncodeToString([]byte(msg))
	nameBase64 := base64.StdEncoding.EncodeToString([]byte(name))
	msgBase64 = url.QueryEscape(msgBase64)
	url := fmt.Sprintf("http://%s:%d/%s/admin/www/notice?name=%s&msg=%s", g.Ip, g.Port, g.Domain, nameBase64, msgBase64)
	return url
}

// 最牛逼英雄信息{q: , heroId: , name}
// name 玩家名字
func (g *GameServerConfig) TopHeroInfo(userId int) string {
	return fmt.Sprintf("http://%s:%d/%s/admin/www/topHeroInfo?userId=%d", g.Ip, g.Port, g.Domain, userId)
}

func (g *GameServerConfig) Payment() string {
	return fmt.Sprintf("http://%s:%d/%s/admin/charge", g.Ip, g.Port, g.Domain)
}

func (g *GameServerConfig) WWWNiubestUserNameUrl(name string) string {
	msgBase64 := base64.StdEncoding.EncodeToString([]byte(name))
	msgBase64 = url.QueryEscape(msgBase64)
	url := fmt.Sprintf("http://%s:%d/%s/admin/www/niubest?msg=%s", g.Ip, g.Port, g.Domain, msgBase64)
	return url
}

// 通知游戏服务器巅峰之夜冠名信息
func (g *GameServerConfig) RankTitleNoticeUrl(typeOfWwa, rank, userId int) string {
	return fmt.Sprintf("http://%s:%d/%s/admin/www/setWwwTitle?type=%d&rank=%d&userId=%d", g.Ip, g.Port, g.Domain,
		typeOfWwa, rank, userId)
}

// 广播
func BrocastNoticeToAllGameServer(msg string) {
	go func() {
		for _, gs := range GameServerList {
			url := gs.NoticeUrl(msg)
			fmt.Println("BrocastNotice : ", url)
			GetGameServer(url)
		}
	}()
}

// 广播并且重复N次
func BrocastNoticeToAllGameServerWithTimeInterval(msg string, times, secInterval int) {
	go func(msg string, times, secInterval int) {
		for i := 0; i < times; i++ {
			BrocastNoticeToAllGameServer(msg)
			time.Sleep(time.Duration(secInterval) * time.Second)
		}
	}(msg, times, secInterval)
}

func BrocastToAllGameServer(fn func(gs *GameServerConfig)) {
	go func() {
		for _, gs := range GameServerList {
			fn(&gs)
		}
	}()
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
