package models

import (
	"fmt"
	"labix.org/v2/mgo/bson"
)

const (
	COL_Behaviour = "behaviours"
	MAX_VIEW_SIZE = 20
)

type Behaviour struct {
	Id       bson.ObjectId `bson:"_id"`
	ZoneId   int           `bson:"zoneid"`   // 服务器编号
	UserId   int           `bson:"userid"`   // 用户
	Name     string        `bson:"name"`     // 名字
	Platform string        `bson:"platform"` // 平台
	Level    int           `bson:"level"`    // 等级
	Exp      int           `bson:"exp"`      // 经验
	Guide    int           `bson:"guide"`    // 指引
	View     []int         `bson:"view"`     // 页面，最多保存30个
	Battle   int           `bson:"battle"`   // 推图进度
}

func NewBehaviour(zoneId, userId int, name, platform string) (err error) {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_Behaviour)
	var count int
	count, err = c.Find(bson.M{"zoneid": zoneId, "userid": userId}).Count()
	if err != nil || count > 0 {
		return
	}
	// 初始状态
	beh := &Behaviour{
		Id:       bson.NewObjectId(),
		ZoneId:   zoneId,
		UserId:   userId,
		Name:     name,
		Platform: platform,
		Level:    1,
		Exp:      0,
		Guide:    0,
		View:     []int{1},
		Battle:   1001,
	}
	return c.Insert(beh)
}

// 引导更新
func GuideChange(zoneId, userId, guide int) (err error) {
	return update(zoneId, userId, bson.M{"guide": guide})
}

// 进入的页面集合
func ViewChange(zoneId, userId, viewId int) (err error) {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_Behaviour)
	beh := &Behaviour{}
	err = c.Find(bson.M{"zoneid": zoneId, "userid": userId}).One(beh)
	if err != nil {
		return
	}
	// 进入页面历史最多保存20个
	beh.View = append(beh.View, viewId)
	if len(beh.View) > MAX_VIEW_SIZE {
		beh.View = beh.View[1:21]
	}
	return c.UpdateId(beh.Id, beh)
}

// 经验更新
func ExpChange(zoneId, userId, level, exp int) (err error) {
	return update(zoneId, userId, bson.M{"level": level, "exp": exp})
}

// 推图进度更新
func BattleChange(zoneId, userId, battleId int) (err error) {
	fmt.Printf("zoneId = %d, userId = %d, battleId = %d \n", zoneId, userId, battleId)
	return update(zoneId, userId, bson.M{"battle": battleId})
}

func update(zoneId, userId int, update interface{}) (err error) {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_Behaviour)
	return c.Update(bson.M{"zoneid": zoneId, "userid": userId}, bson.M{"$set": update})
}
