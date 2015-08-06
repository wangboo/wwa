package controllers

import (
	"github.com/revel/revel"
	"github.com/wangboo/wwa/app/models"
	// "labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

type WWAWeekCtrl struct {
	*revel.Controller
}

const (
	TYPE_WWW_VIEW_NORMAL    = 0 // 周一到周六显示普通挑战积分（周一至周六，周六11点结算常规赛积分）
	TYPE_WWW_VIEW_WAIT      = 1 // 准备界面11点到12点，不能点击
	TYPE_WWW_VIEW_BET       = 2 // 周日下注界面（周六23点到周日0点）
	TYPE_WWW_VIEW_FIGHT_IN  = 3 // 周日挑战界面（周日20点-21点）参与人员
	TYPE_WWW_VIEW_FIGHT_OUT = 4 // 周日挑战界面（周日20点-21点）非参与人员
	TYPE_WWW_VIEW_OVER      = 5 // 周日挑战结果界面（周日21点-周日0点）
)

// 首页
// 根据日期来判断现在应该展示何种界面
func (w *WWAWeekCtrl) MainPage(zoneId, userId, typeOfWwa int) revel.Result {
	now := time.Now()
	weekday := now.Weekday()
	hour := now.Hour()
	// test
	// weekday, hour = 0, 10
	switch weekday {
	case 0:
		// 周天
		if hour < 19 {
			// 下注界面
			return w.ShowTop20(zoneId, userId, typeOfWwa, TYPE_WWW_VIEW_BET)
		} else {
			// 挑战界面
			return w.ShowTop20(zoneId, userId, typeOfWwa, TYPE_WWW_VIEW_FIGHT_IN)
		}
	case 6:
		if hour > 23 {
			// 周六 提示：巅峰之夜准备中，请在周日0点以后查看
			return w.RenderJson(FailWithMsg("巅峰之夜准备中，请在周日0点以后查看"))
		} else {
			// 显示
			return w.ShowTop20(zoneId, userId, typeOfWwa, TYPE_WWW_VIEW_NORMAL)
		}
	default:
		// 周一到周五
		return w.ShowTop20(zoneId, userId, typeOfWwa, TYPE_WWW_VIEW_NORMAL)
	}
	return w.RenderJson(Succ())
}

// 周一到周六显示巅峰竞技排行
func (w *WWAWeekCtrl) ShowTop20(zoneId, userId, typeOfWwa, typeOfView int) revel.Result {
	selfWeek, err := models.FindWWAWeekByZoneIdAndUserId(zoneId, userId)
	if err != nil {
		return w.RenderJson(FailWithError(err))
	}
	selfWWA, err := models.FindWWAInRedis(zoneId, userId)
	if err != nil {
		return w.RenderJson(FailWithError(err))
	}
	revel.INFO.Println("selfWWA.Type() = ", selfWWA.Type())
	if typeOfWwa == -1 {
		typeOfWwa = selfWWA.Type()
	}
	top20 := models.UserWWAWeekTop20(typeOfWwa)
	rst := Succ()
	rst["score"] = selfWeek.Score
	rst["typeOfWwa"] = typeOfWwa
	rst["typeOfView"] = typeOfView
	rst["typeOfMine"] = selfWWA.Type()
	// 如果积分为0，那么名次为-1
	if selfWeek.Score > 0 {
		rst["rank"] = models.RankInWeekWWA(selfWeek.Score, selfWeek.Pow, selfWeek.Type)
	} else {
		rst["rank"] = -1
	}
	list := []map[string]interface{}{}
	for index, week := range top20 {
		info := bson.M{}
		item, err := models.FindWWAInRedis(week.ZoneId, week.UserId)
		if err != nil {
			continue
		}
		info["id"] = week.Id.Hex()
		info["zoneId"] = item.ZoneId()
		info["userId"] = item.UserId()
		info["frame"] = item.Frame()
		info["img"] = item.Img()
		info["pow"] = item.Pow()
		info["rank"] = index + 1
		info["zoneName"] = item.ZoneName()
		info["name"] = item.Name()
		info["level"] = item.Level()
		switch typeOfView {
		case TYPE_WWW_VIEW_NORMAL:
			info["score"] = week.Score
		case TYPE_WWW_VIEW_BET:
			info["score"] = week.Score
			// 对他的下注金额
			info["gold"] = models.FindUserBetInUserSum(zoneId, userId, week.Id)
		case TYPE_WWW_VIEW_FIGHT_IN:
			info["score"] = week.PlayoffScore
		case TYPE_WWW_VIEW_OVER:
			info["score"] = week.PlayoffScore
		}
		list = append(list, info)
	}
	if typeOfView == TYPE_WWW_VIEW_BET {
		rst["bet"] = models.FindUserBetSum(zoneId, userId)
	}
	rst["list"] = list
	return w.RenderJson(rst)
}

// 获得下一位挑战者信息
func (w *WWAWeekCtrl) NextFightUser(zoneId, userId int) revel.Result {
	week, err := models.FindWWAWeekByZoneIdAndUserId(zoneId, userId)
	if err != nil {
		return w.RenderJson(FailWithError(err))
	}
	if len(week.WaitList) == 0 {
		// 已经打完了
		return w.RenderJson(Succ(""))
	}
	next, err := models.FindWWAWeekById(week.FightingId)
	if err != nil {
		return w.RenderJson(FailWithError(err))
	}
	rst := Succ()
	rst["zoneId"] = next.ZoneId
	rst["userId"] = next.UserId
	return w.RenderJson(rst)
}

// 挑战结果，无论挑战成功失败都将跳到下一位挑战者
// rst 0胜1负2平  胜利3分，打平1分，失败0分
func (w *WWAWeekCtrl) FightResult(zoneId, userId, rst int) revel.Result {
	week, err := models.FindWWAWeekByZoneIdAndUserId(zoneId, userId)
	if err != nil {
		return w.RenderJson(FailWithError(err))
	}
	score := 0
	switch rst {
	case 0:
		score = 3
	case 2:
		score = 1
	}
	fightResult := models.CreateUserWeekWWAFightResult(week.FightingId, rst, score)
	week.FightedList = append(week.FightedList, *fightResult)
	// 切换到下一个挑战者
	if len(week.WaitList) > 0 {
		week.FightingId = week.WaitList[0]
		week.WaitList = week.WaitList[1:]
	}
	// save
	week.Save()
	return w.RenderJson(Succ())
}
