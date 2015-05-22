package models

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Rank struct {
	BaseModel
	UserId int    `json:"id"`        // 玩家id
	Score  int    `sql:"default:0"`  // 竞技场积分
	Level  int    `json:"level"`     // 等级
	Name   string `json:"nick_name"` // 昵称
	Hero   int    `json:"hero"`      // 头像
	Q      int    `json:"q"`         // 头像品质
	Pow    int    `json:"pow"`       // 战斗力
	ZoneId int    `sql:"not null"`   // 游戏服务器id
	Type   int    `sql:"not null"`   // 排行榜类别 分为3中竞技场 0低等级（15-34），1中等级（35-54），2高等级（55-）
	Img    int    `json:"img"`
	Frame  int    `json:"frame"`
	// 玩家竞技场数据
	Data     string
	ZoneName string
}

func (r *Rank) String() string {
	return fmt.Sprintf("UserId=%d,Level=%d,Name=%s,Pow=%d,Type= \n", r.UserId, r.Level, r.Name, r.Pow, r.Type)
}

const RANK_SCORE_SUB = 100000

// 存到排名redis的值
func (r *Rank) ToDetailKey() string {
	return fmt.Sprintf("%d,%d,%d,%s,%d,%d,%d,%d,%s,%d,%d,%d", r.UserId, r.Score, r.Level, r.Name, r.Pow, r.Hero, r.Q, r.ZoneId, r.ZoneName, r.Type,
		r.Img, r.Frame)
}

// 存到redis排名的 SortedSet 的table名
func (r *Rank) ToRedisRankName() string {
	return fmt.Sprintf("wwa_%d", r.Type)
}

func (r *Rank) ToSimpleKey() string {
	return ToSimpleKey(r.ZoneId, r.UserId)
}

// 存到redis hash基本信息中的key(k-v : ToSimpleKey()-ToDetailKey())
func ToSimpleKey(zoneId, userId int) string {
	return fmt.Sprintf("%d,%d", zoneId, userId)
}

// 队伍信息键 set gi_1,1 ... 60 存放队伍信息
func GroupInfoKey(zoneId, userId int) string {
	return fmt.Sprintf("gi_%d,%d", zoneId, userId)
}

// 幕僚信息键 set gi_1,1 ... 60 存放幕僚信息
func MuInfoKey(zoneId, userId int) string {
	return fmt.Sprintf("mi_%d,%d", zoneId, userId)
}

// 队伍信息键 set gd_1,1 ... 60 存放队伍战斗信息
func GroupDataKey(zoneId, userId int) string {
	return fmt.Sprintf("gd_%d,%d", zoneId, userId)
}

// 从游戏服务器获取队伍信息
func GetGroupInfoFromGameServer(zoneId, userId int) string {
	url := FindGameServer(zoneId).GroupInfoUrl(userId)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("GetGroupInfoFromGameServer: 服务器%s无法访问", url)
		return ""
	}
	str, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("GetGroupInfoFromGameServer: %s应答无法读取", url)
		return ""
	}
	return fmt.Sprintf("%s", str)
}

// 从游戏服务器获取幕僚信息
func GetMuInfoFromGameServer(zoneId, userId int) string {
	url := FindGameServer(zoneId).MuInfoUrl(userId)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("GetGroupInfoFromGameServer: 服务器%s无法访问", url)
		return ""
	}
	str, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("GetGroupInfoFromGameServer: %s应答无法读取", url)
		return ""
	}
	return fmt.Sprintf("%s", str)
}

// 从游戏服务器获取队伍信息
func GetGroupDataFromGameServer(zoneId, userId int) string {
	url := FindGameServer(zoneId).GroupDataUrl(userId)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("GetGroupInfoFromGameServer: 服务器%s无法访问", url)
		return ""
	}
	str, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("GetGroupInfoFromGameServer: %s应答无法读取", url)
		return ""
	}
	return fmt.Sprintf("%s", str)
}
