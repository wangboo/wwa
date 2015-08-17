package controllers

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/revel/revel"
	"github.com/wangboo/wwa/app/jobs"
	"github.com/wangboo/wwa/app/models"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	// "time"
)

type WWAWeekCtrl struct {
	*revel.Controller
}

// 首页
// 根据日期来判断现在应该展示何种界面
func (w *WWAWeekCtrl) MainPage(zoneId, userId, typeOfWwa int) revel.Result {
	// 结果
	var resp map[string]interface{}
	var week *models.UserWWAWeek
	var err error
	typeOfView := models.GetSysWWAWeekState()
	// 获取玩家巅峰之夜数据
	switch typeOfView {
	case models.TYPE_WWW_VIEW_FIGHT_IN:
		week, err = models.FindWWAWeekByZoneIdAndUserId(zoneId, userId)
		if err != nil {
			if err == mgo.ErrNotFound {
				// 找不到玩家数据
				typeOfView = models.TYPE_WWW_VIEW_FIGHT_OUT
			} else {
				return w.RenderJson(FailWithError(err))
			}
		}
		if !week.InTop20 {
			typeOfView = models.TYPE_WWW_VIEW_FIGHT_OUT
		}
	default:
		week, err = models.FindOrCreateWWAWeekByZoneIdAndUserId(zoneId, userId)
		if err != nil {
			return w.RenderJson(FailWithError(err))
		}
	}
	// 判断当前状态
	switch typeOfView {
	case models.TYPE_WWW_VIEW_FIGHT_IN:
		if week.IsFinish {
			// 挑战完毕后会现实积分界面
			resp = w.ShowTop20(week, zoneId, userId, typeOfWwa, models.TYPE_WWW_VIEW_FIGHT_OUT)
		} else {
			resp = w.FightInView(week, zoneId, userId, typeOfWwa)
		}
	case models.TYPE_WWW_VIEW_PREPARE:
		resp = FailWithMsg("巅峰之夜准备中，请在周日0点以后查看")
	case models.TYPE_WWW_VIEW_OVER:
		selfWWA, err := models.FindWWAInRedis(zoneId, userId)
		if err != nil {
			return w.RenderJson(FailWithError(err))
		}
		if typeOfWwa == -1 {
			typeOfWwa = selfWWA.Type()
		}
		sys := models.FindSysWWAWeek()
		rst := Succ()
		rst["typeOfView"] = models.TYPE_WWW_VIEW_OVER
		rst["typeOfWwa"] = typeOfWwa
		rst["typeOfMine"] = selfWWA.Type()
		rst["open"] = sys.IsPlayoffOn[typeOfWwa]
		if sys.IsPlayoffOn[typeOfWwa] {
			rst["list"] = sys.Top3Cache[typeOfWwa]
		}
		if week.PlayoffScore > 0 {
			rst["rank"] = models.RankInWeekWWA(week.Score, week.Pow, week.Type)
			rst["score"] = week.PlayoffScore
		} else {
			rst["rank"] = -1
			rst["score"] = 0
		}
		return w.RenderJson(rst)
	default:
		resp = w.ShowTop20(week, zoneId, userId, typeOfWwa, typeOfView)
	}
	return w.RenderJson(resp)
}

// 巅峰竞技，挑战界面
func (w *WWAWeekCtrl) FightInView(week *models.UserWWAWeek, zoneId, userId, typeOfWwa int) map[string]interface{} {
	rst := Succ()
	// score, pow, typeOfWwa
	rst["open"] = true
	rst["typeOfView"] = models.TYPE_WWW_VIEW_FIGHT_IN
	rst["typeOfMine"] = week.Type
	rst["rank"] = models.RankInWeekWWA(week.PlayoffScore, week.Pow, week.Type)
	rst["score"] = week.PlayoffScore
	fightedSize := len(week.FightedList)
	waitSize := len(week.WaitList)
	max := fightedSize + waitSize
	rst["cur"] = fightedSize
	if week.FightingId.Valid() {
		max += 1
	}
	rst["mx"] = max
	// 积分
	win, lose, deuce := 0, 0, 0
	for _, item := range week.FightedList {
		switch item.FightRst {
		case 0:
			win += 1
		case 1:
			lose += 1
		case 2:
			deuce += 1
		}
	}
	rst["score"], rst["win"], rst["deuce"], rst["lose"] = week.PlayoffScore, win, deuce, lose
	fightWeek, err := models.FindWWAWeekById(week.FightingId)
	if err == nil {
		wwa, err := models.FindWWAInRedis(fightWeek.ZoneId, fightWeek.UserId)
		if err == nil {
			rst["pow"] = wwa.Pow()
			rst["name"] = wwa.Name()
			rst["lev"] = wwa.Level()
			rst["zoneName"] = wwa.ZoneName()
			rst["frame"] = wwa.Frame()
			rst["img"] = wwa.Img()
			rst["aid"] = fightWeek.ZoneId
			rst["uid"] = fightWeek.UserId
		}
	}
	return rst
}

// 周一到周六显示巅峰竞技排行
func (w *WWAWeekCtrl) ShowTop20(selfWeek *models.UserWWAWeek, zoneId, userId, typeOfWwa, typeOfView int) map[string]interface{} {
	selfWWA, err := models.FindWWAInRedis(zoneId, userId)
	if err != nil {
		return FailWithError(err)
	}
	revel.INFO.Println("selfWWA.Type() = ", selfWWA.Type())
	if typeOfWwa == -1 {
		typeOfWwa = selfWWA.Type()
	}
	var top20 []models.UserWWAWeek
	if typeOfView == models.TYPE_WWW_VIEW_FIGHT_IN ||
		typeOfView == models.TYPE_WWW_VIEW_FIGHT_OUT ||
		typeOfView == models.TYPE_WWW_VIEW_OVER {
		top20 = models.UserWWAWeekTop20(typeOfWwa)
	} else {
		top20 = models.UserWWAWeekTop20ByScore(typeOfWwa)
	}
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
	// 开赛条件检查
	showDetail := true
	if typeOfView == models.TYPE_WWW_VIEW_BET_ENABLE ||
		typeOfView == models.TYPE_WWW_VIEW_BET_DISABLE ||
		typeOfView == models.TYPE_WWW_VIEW_FIGHT_OUT ||
		typeOfView == models.TYPE_WWW_VIEW_OVER {
		sys := models.FindSysWWAWeek()
		showDetail = sys.IsPlayoffOn[typeOfWwa]
		rst["open"] = showDetail
	}
	// 满足开赛条件
	if showDetail {
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
			info["type"] = item.Type()
			info["name"] = item.Name()
			info["level"] = item.Level()
			switch typeOfView {
			case models.TYPE_WWW_VIEW_NORMAL:
				info["score"] = week.Score
			case models.TYPE_WWW_VIEW_BET_DISABLE:
				info["score"] = 0
			case models.TYPE_WWW_VIEW_BET_ENABLE:
				info["score"] = week.Score
				// 对他的下注金额
				info["gold"] = models.FindUserBetInUserSum(zoneId, userId, week.Id)
			case models.TYPE_WWW_VIEW_FIGHT_IN:
				info["score"] = week.PlayoffScore
			case models.TYPE_WWW_VIEW_FIGHT_OUT:
				info["score"] = week.PlayoffScore
			case models.TYPE_WWW_VIEW_OVER:
				info["score"] = week.PlayoffScore
			}
			list = append(list, info)
		}
	}

	if typeOfView == models.TYPE_WWW_VIEW_BET_ENABLE {
		rst["bet"] = models.FindUserBetSum(zoneId, userId)
	}
	rst["list"] = list
	return rst
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
	lenOfFightedList := len(week.FightedList)
	if rst == 0 {
		// 连胜次数
		lsTimes := 1
		for i := lenOfFightedList - 1; i >= 0; i-- {
			if week.FightedList[i].FightRst > 0 {
				break
			}
			lsTimes++
		}
		if lsTimes >= 3 {
			// 广播连胜消息
			wwa, err := models.FindWWAInRedis(zoneId, userId)
			if err == nil {
				models.BrocastNoticeToAllGameServer(fmt.Sprintf("玩家%s在巅峰之夜%s段位的对决中大杀四方，无人可挡，已经%d连胜啦！",
					wwa.Name(), models.WwaTypeToName(week.Type), lsTimes))
			}
		}
	}
	fightResult := models.CreateUserWeekWWAFightResult(week.FightingId, rst, score)
	week.FightedList = append(week.FightedList, *fightResult)
	// 切换到下一个挑战者
	if len(week.WaitList) > 0 {
		week.FightingId = week.WaitList[0]
		week.WaitList = week.WaitList[1:]
	} else {
		week.IsFinish = true
	}
	week.PlayoffScore += score
	// save
	week.Save()
	return w.RenderJson(Succ("score", score))
}

// 挑战对象玩家的基本信息
func (w *WWAWeekCtrl) FightInfo(zoneId, userId int) revel.Result {
	week, err := models.FindWWAWeekByZoneIdAndUserId(zoneId, userId)
	if err != nil {
		return w.RenderJson(FailWithError(err))
	}
	if !week.FightingId.Valid() {
		return w.RenderJson(FailWithMsg("找不到对方玩家信息"))
	}
	fweek, err := models.FindWWAWeekById(week.FightingId)
	if err != nil {
		return w.RenderJson(FailWithError(err))
	}
	wwa, err := models.FindWWAInRedis(fweek.ZoneId, fweek.UserId)
	if err != nil {
		return w.RenderJson(FailWithError(err))
	}
	rst := Succ("uid", fweek.UserId, "aid", fweek.ZoneId, "lev", wwa.Level())
	return w.RenderJson(rst)
}

// 挑战
func (w *WWAWeekCtrl) Fight(zoneId, userId int) revel.Result {
	week, err := models.FindWWAWeekByZoneIdAndUserId(zoneId, userId)
	if err != nil {
		return w.RenderJson(FailWithError(err))
	}
	if !week.FightingId.Valid() {
		return w.RenderJson(FailWithMsg("找不到对方玩家信息"))
	}
	fweek, err := models.FindWWAWeekById(week.FightingId)
	if err != nil {
		return w.RenderJson(FailWithError(err))
	}
	return w.RenderText(GetGroupData(fweek.ZoneId, fweek.UserId))
}

// 开赛手动触发定时器
func (w *WWAWeekCtrl) BeginFight() revel.Result {
	beginJob := mjob.WWAWeekFightBeginJob{}
	beginJob.Run()
	return w.RenderJson(Succ())
}

// 更换临时巅峰竞技类型
func (w *WWAWeekCtrl) ChangeType(zoneId, userId, typeOfWwa, lev int) revel.Result {
	state := models.GetSysWWAWeekState()
	// 只有在非开启巅峰之夜模式下才能变换
	if state != models.TYPE_WWW_VIEW_NORMAL {
		return w.RenderJson(Succ())
	}
	week, err := models.FindWWAWeekByZoneIdAndUserId(zoneId, userId)
	if err != nil {
		return w.RenderJson(FailWithError(err))
	}
	week.Type = typeOfWwa
	week.Save()
	wwa, err := models.FindWWAInRedis(zoneId, userId)
	if err != nil {
		return w.RenderJson(FailWithError(err))
	}
	wwa.SetLevel(lev)
	wwa.SetType(typeOfWwa)
	wwa.UpdateToRedis()
	return w.RenderJson(Succ())
}

// 全游戏广播
func (w *WWAWeekCtrl) NoticeOnTV(msg string, repeat bool, times, sec int) revel.Result {
	if !repeat {
		models.BrocastNoticeToAllGameServer(msg)
	} else {
		models.BrocastNoticeToAllGameServerWithTimeInterval(msg, times, sec)
	}
	return w.RenderJson(Succ())
}

// 最牛逼的玩家名
func (w *WWAWeekCtrl) Niubest() revel.Result {
	cli := models.RedisPool.Get()
	defer cli.Close()
	name, err := redis.String(cli.Do("GET", "wwa_3_first_rank"))
	if err != nil {
		return w.RenderJson(Succ("name", ""))
	}
	return w.RenderJson(Succ("name", name))
}
