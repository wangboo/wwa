package models

import (
	"fmt"
)

type Rank struct {
	BaseModel
	UserId int    `json:"id"`        // 玩家id
	Score  int    `sql:"default:0"`  // 竞技场积分
	Level  int    `json:"level"`     // 等级
	Name   string `json:"nick_name"` // 昵称
	Pow    int    `json:"pow"`       // 战斗力
	ZoneId int    `sql:"not null"`   // 游戏服务器id
	Type   int    `sql:"not null"`   // 排行榜类别 分为3中竞技场 0低等级（15-34），1中等级（35-54），2高等级（55-）
	// 玩家竞技场数据
	Data string
}

func FindRankById(zoneId, id, t int) *Rank {
	rank := &Rank{ZoneId: zoneId, UserId: id, Type: t}
	DB.Find(rank)
	return rank
}

// 存到排名redis的值
func (r *Rank) ToRedisRankValue() string {
	return fmt.Sprintf("%d,%d,%d,%s,%d,%d,%d", r.UserId, r.Score, r.Level, r.Name, r.Pow, r.ZoneId, r.Type)
}

// 存到redis排名的SortedSet 的table名
func (r *Rank) ToRedisRankName() string {
	return fmt.Sprintf("wwa_%d", r.Type)
}
