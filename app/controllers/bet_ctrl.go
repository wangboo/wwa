package controllers

import (
	"github.com/revel/revel"
	"github.com/wangboo/wwa/app/models"
	"labix.org/v2/mgo/bson"
)

type BetCtrl struct {
	*revel.Controller
}

// 下注 zoneId 玩家区服，useId 玩家id，weekId 下注目标玩家, gold 元宝
func (b *BetCtrl) Bet(zoneId, userId, gold int, weekId string) revel.Result {
	if !bson.IsObjectIdHex(weekId) {
		return b.RenderJson(FailWithMsg("weekId错误不是ObjectId"))
	}
	weekObjectId := bson.ObjectIdHex(weekId)
	_, err := models.BetTo(zoneId, userId, gold, weekObjectId)
	if err != nil {
		return b.RenderJson(FailWithError(err))
	}
	return b.RenderJson(Succ())
}

// 总和
func (b *BetCtrl) Sum() revel.Result {
	sum := models.FindBetSum()
	return b.RenderJson(Succ("sum", sum))
}
