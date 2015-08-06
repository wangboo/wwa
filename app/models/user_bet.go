package models

import (
	"fmt"
	"github.com/revel/revel"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

// 玩家下注
type UserBet struct {
	Id        bson.ObjectId `bson:"_id"`
	ZoneId    int           `bson:"zone_id"`     // 下注玩家区服
	UserId    int           `bson:"user_id"`     // 下注玩家id
	Type      int           `bson:"type"`        // 下注跨服竞技类别
	BetUserId bson.ObjectId `bson:"bet_user_id"` // 投注给玩家，UserWWAWeek 表的Id
	Gold      int           `bson:"gold"`        // 下注金额
}

const (
	COL_USER_BET = "user_bets"
)

// 下注给指定玩家
func BetTo(zoneId, userId, gold int, betUserId bson.ObjectId) (bet *UserBet, err error) {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_BET)
	userBet, err := FindWWAWeekById(betUserId)
	if err != nil {
		return nil, fmt.Errorf("找不到下注的玩家id")
	}
	bet = &UserBet{}
	err = c.Find(bson.M{"zone_id": zoneId, "user_id": userId}).Select(bson.M{"_id": 1, "gold": 1}).One(&bet)
	if err != nil {
		if err == mgo.ErrNotFound {
			bet.Id = bson.NewObjectId()
			bet.ZoneId = zoneId
			bet.UserId = userId
			bet.Type = userBet.Type
			bet.BetUserId = betUserId
			bet.Gold = gold
			c.Insert(&bet)
			err = nil
		} else {
			// 查询错误
			revel.ERROR.Println("查询出现错误：", err)
			return
		}
	} else {
		bet.Gold += gold
		c.UpdateId(bet.Id, bson.M{"$set": bson.M{"gold": bet.Gold}})
	}
	return
}

// 查询玩家总下注
func FindUserBetSum(zoneId, userId int) int {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_BET)
	rst := bson.M{}
	err := c.Pipe([]bson.M{
		{"$match": bson.M{"zone_id": zoneId, "user_id": userId}},
		{"$group": bson.M{"_id": "", "totle": bson.M{"$sum": "$gold"}}},
	}).One(rst)
	sum := 0
	if err != nil {
		revel.ERROR.Println("error ", err)
	} else {
		sum = int(rst["totle"].(float64))
	}
	return sum
}

// 查询玩家对某一个玩家的下注总和
func FindUserBetInUserSum(zoneId, userId int, betUserId bson.ObjectId) int {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_BET)
	rst := bson.M{}
	err := c.Pipe([]bson.M{
		{"$match": bson.M{"zone_id": zoneId, "user_id": userId, "bet_user_id": betUserId}},
		{"$group": bson.M{"_id": "", "totle": bson.M{"$sum": "$gold"}}},
	}).One(rst)
	sum := 0
	if err != nil {
		revel.ERROR.Println("error ", err)
	} else {
		sum = rst["sum"].(int)
	}
	return sum
}

// 查询总下注金额
func FindBetSum() int {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_BET)
	rst := bson.M{}
	err := c.Pipe([]bson.M{
		{"$group": bson.M{"_id": "", "totle": bson.M{"$sum": "$gold"}}},
	}).One(&rst)
	sum := 0
	if err != nil {
		revel.ERROR.Println("error ", err)
	} else {
		sum = rst["totle"].(int)
	}
	return sum
}

// 查询总下注金额
func FindBetSumByType(typeOfWwa int) int {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_BET)
	rst := bson.M{}
	err := c.Pipe([]bson.M{
		{"$match": bson.M{"type": typeOfWwa}},
		{"$group": bson.M{"_id": "", "totle": bson.M{"$sum": "$gold"}}},
	}).One(&rst)
	sum := 0
	if err != nil {
		revel.ERROR.Println("error ", err)
	} else {
		sum = rst["totle"].(int)
	}
	return sum
}

// 查询该难度冠军的押注回报比率
func FindBetRateByType(typeOfWwa int, winerId bson.ObjectId) float64 {
	return 0
}

// 查询玩家下注给指定玩家的集合
func FindUserBetOnUser(betUserId bson.ObjectId) (list []UserBet) {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_BET)
	list = []UserBet{}
	c.Find(bson.M{"bet_user_id": betUserId}).All(&list)
	return
}

// 寻找该跨服竞技类型的所有押注
func FindUserBetByType(typeOfWwa int) (list []UserBet) {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_BET)
	list = []UserBet{}
	c.Find(bson.M{"type": typeOfWwa}).All(&list)
	return
}

// 查询玩家的总下注情况
// return [{"_id": {"zone_id": , "user_id": }, "gold": }]
func FindUserBetSumedByZoneAndUser(typeOfWwa int) (list []map[string]interface{}) {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_BET)
	list = []map[string]interface{}{}
	c.Pipe([]bson.M{
		{"$group": bson.M{
			"_id":  bson.M{"zone_id": "$zone_id", "user_id": "$user_id"},
			"gold": bson.M{"$sum": "$gold"},
		}},
	}).All(&list)
	return
}
