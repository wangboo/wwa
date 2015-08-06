package models

import (
	"github.com/revel/revel"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

// 玩家跨服竞技周积分
type UserWWAWeek struct {
	Id           bson.ObjectId   `bson:"_id"`                   // id
	ZoneId       int             `bson:"zone_id"`               // 玩家服务器id
	UserId       int             `bson:"user_id"`               // 玩家id
	Score        int             `bson:"score"`                 // 周1-6总积分
	Type         int             `bson:"type"`                  // 挑战级别
	Pow          int             `bson:"pow"`                   // 战斗力
	FightedList  []FightResult   `bson:"win_list"`              // 挑战胜利列表
	WaitList     []bson.ObjectId `bson:"wait_list"`             // 未挑战列表
	FightingId   bson.ObjectId   `bson:"fighting_id,omitempty"` // 正在挑战的玩家id
	PlayoffScore int             `bson:"playoff_score"`         // 季后赛积分
	CreatedAt    time.Time       `bson:"created_at"`            // 创建时间
	IsSend       bool            `bson:"is_send"`               // 是否发放奖励
}

// 挑战结果
type FightResult struct {
	UserId   bson.ObjectId `bson:"user_id"`   // 玩家id（UserWWAWeek主键）
	FightRst int           `bson:"fight_rst"` // 挑战结果0胜1负2平
	Score    int           `bson:"score"`     // 获得分数
}

const (
	COL_USER_WWA_WEEK        = "user_wwa_weeks"        // 当前巅峰之夜记录表
	COL_USER_WWA_WEEK_RECORD = "user_wwa_week_records" // 历史表
)

func CreateUserWeekWWAFightResult(fightId bson.ObjectId, rst, score int) *FightResult {
	return &FightResult{
		UserId:   fightId,
		FightRst: rst,
		Score:    score,
	}
}

// 通过zoneId和userId查询玩家周巅峰之夜数据
func FindWWAWeekByZoneIdAndUserId(zoneId, userId int) (week *UserWWAWeek, err error) {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_WWA_WEEK)
	week = &UserWWAWeek{}
	err = c.Find(bson.M{"zone_id": zoneId, "user_id": userId}).One(week)
	if err != nil {
		if err == mgo.ErrNotFound {
			wwa, redisErr := FindWWAInRedis(zoneId, userId)
			if redisErr != nil {
				revel.ERROR.Println("redis error:", redisErr)
				return nil, redisErr
			}
			week.Id = bson.NewObjectId()
			week.ZoneId = zoneId
			week.UserId = userId
			week.Pow = wwa.Pow()
			week.Score = wwa.Score()
			week.Type = wwa.Type()
			week.CreatedAt = time.Now()
			week.FightedList = []FightResult{}
			week.WaitList = []bson.ObjectId{}
			err = c.Insert(week)
			revel.INFO.Println("week.Id = ", week.Id)
			if err != nil {
				revel.ERROR.Println("save wwa_week err: ", err)
			}
			err = nil
		} else {
			// mongo error
			revel.ERROR.Printf("FindWWAWeekByZoneIdAndUserId error : %s \n", err.Error())
			return
		}
	}
	return
}

// 通过玩家巅峰之夜数据id查询该玩家信息
func FindWWAWeekById(Id bson.ObjectId) (week *UserWWAWeek, err error) {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_WWA_WEEK)
	week = &UserWWAWeek{}
	err = c.FindId(Id).One(week)
	return
}

// 增加或减少积分 return true if success
func UserWWAWeekScoreChange(zoneId, userId, score, typeOfWwa int) bool {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_WWA_WEEK)
	week, err := FindWWAWeekByZoneIdAndUserId(zoneId, userId)
	if err != nil {
		revel.ERROR.Printf("update failed: %s \n", err.Error())
		return false
	}
	newScore := week.Score + score
	c.UpdateId(week.Id, bson.M{"$set": bson.M{"score": newScore}})
	return true
}

// 周排名前20
func UserWWAWeekTop20(typeOfWwa int) (list []UserWWAWeek) {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_WWA_WEEK)
	list = []UserWWAWeek{}
	c.Find(bson.M{"type": typeOfWwa, "score": bson.M{"$gt": 0}}).Sort("-score", "-pow").Limit(20).All(&list)
	return
}

// 巅峰之夜20名单
func UserWWAWeekPlayoff(typeOfWwa int) (list []UserWWAWeek) {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_WWA_WEEK)
	list = []UserWWAWeek{}
	c.Find(nil).Sort("-playoff_score", "-pow").Limit(20).All(&list)
	return
}

// 切换到季后赛模式，删除20名以后的人员(指定类型)
func UserWWAWeekSwitch2PlayoffByType(typeOfWwa int) {
	list := UserWWAWeekPlayoff(typeOfWwa)
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_WWA_WEEK)
	c.RemoveAll(bson.M{"type": typeOfWwa})
	for _, week := range list {
		c.Insert(week)
	}
}

// 切换到季后赛模式，删除20名以后的人员
func UserWWAWeekSwitch2Playoff() {
	UserWWAWeekSwitch2PlayoffByType(0)
	UserWWAWeekSwitch2PlayoffByType(1)
	UserWWAWeekSwitch2PlayoffByType(2)
}

// 重置当前等级的巅峰之夜数据
func ResetScore(typeOfWwa int) {
	s := Session()
	defer s.Close()
	colRec := s.DB(DB_NAME).C(COL_USER_WWA_WEEK_RECORD)
	list := UserWWAWeekTop20(typeOfWwa)
	for _, r := range list {
		r.CreatedAt = time.Now()
		colRec.Insert(&r)
	}
	colWeek := s.DB(DB_NAME).C(COL_USER_WWA_WEEK)
	colWeek.RemoveAll(bson.M{"type": typeOfWwa})
}

// 找到积分大于score或者(积分等于score且战斗力大于pow)的所有玩家的数量
func RankInWeekWWA(score, pow, typeOfWwa int) int {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_WWA_WEEK)
	// db.test.find({"$or":[{score: {"$gt": 12}},{score: 12, pow: {"$gt": 11}}]})
	findExp := bson.M{
		"type":  typeOfWwa,
		"score": bson.M{"$gt": 0},
		"$or": []map[string]interface{}{
			bson.M{"score": bson.M{"$gt": score}},
			bson.M{"score": score, "pow": bson.M{"$gt": pow}},
		}}
	count, _ := c.Find(findExp).Count()
	return count + 1
}

// 保存
func (u *UserWWAWeek) Save() {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_WWA_WEEK)
	c.UpdateId(u.Id, u)
}
