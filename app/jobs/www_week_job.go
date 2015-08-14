package mjob

import (
	"fmt"
	"github.com/revel/revel"
	"github.com/wangboo/wwa/app/models"
	"time"
)

// 挑战前初始化定时器
type WWAWeekFightBeginJob struct {
}

// 挑战结束，发奖定时器
type WWWWeekFightEndJob struct {
}

// ------------------- 挑战前初始化定时器 -------------------

func (w *WWAWeekFightBeginJob) Run() {
	defer catchException()
	FightBeginImpl()
}

// 开始
func FightBeginImpl() {
	models.UserWWAWeekSwitch2Playoff()
}

// ------------------- 挑战结束定时器 -------------------

// 挑战结束，重置定时器

// 定时器
func (w *WWWWeekFightEndJob) Run() {
	defer catchException()
	FightEndImpl()
}

// 挑战结束，发奖
func FightEndImpl() {
	models.WwaTypeForeach(FightEndByType)
}

func FightEndByType(typeOfWwa int) {
	// 挑战者奖励
	list := models.UserWWAWeekPlayoff(typeOfWwa)
	for index, item := range list {
		rank := index + 1
		reward := models.BaseWWAWeekRewardGetRewardByTypeAndRank(typeOfWwa, rank)
		_, week := time.Now().ISOWeek()
		msg := fmt.Sprintf("恭喜您在第%d周跨服巅峰之夜中荣获第%d名，请点击领取奖励。", week, rank)
		resp := sendMail(item.ZoneId, item.UserId, msg, reward)
		if resp == "ok" {
			item.IsSend = true
			item.Save()
		} else {
			revel.ERROR.Println("定时任务WWWWeekFightEndJob发放奖励出现错误: ", resp)
		}
	}

	if len(list) == 0 {
		revel.WARN.Println("没有胜利者，无法发放奖励")
		// 退回所有玩家押注元宝
		sendBackToUserBetByType(typeOfWwa)
		return
	}
	winner := &list[0]
	// sum := models.FindBetSumByType(typeOfWwa)
	betList := models.FindUserBetByType(typeOfWwa)
	// 下注奖励
	for _, bet := range betList {
		if bet.BetUserId == winner.Id {
			// 中奖

		} else {
			// 没中
		}
	}
}

func sendMail(zoneId, userId int, msg, reward string) string {
	gs := models.FindGameServer(zoneId)
	byteOfResp, err := models.GetGameServer(gs.CommonRewardMail(userId, msg, reward))
	if err != nil {
		return "fail"
	}
	return string(byteOfResp)
}

// 退回跨服竞技类型为typeOfWwa的全部押注元宝
func sendBackToUserBetByType(typeOfWwa int) {
	list := models.FindUserBetSumedByZoneAndUser(typeOfWwa)
	for _, item := range list {
		id := item["_id"].(map[string]interface{})
		userId := id["user_id"].(int)
		zoneId := id["zone_id"].(int)
		gold := item["gold"].(int)
		msg := fmt.Sprintf("由于本次巅峰之夜%s境界没有第一名，退回您的押注元宝。", models.WwaTypeToName(typeOfWwa))
		reward := fmt.Sprintf("g-%d", gold)
		gs := models.FindGameServer(zoneId)
		models.GetGameServer(gs.CommonRewardMail(userId, msg, reward))
	}
}
