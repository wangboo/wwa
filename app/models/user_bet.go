package models

import (
	"encoding/json"
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
	err = c.Find(bson.M{"zone_id": zoneId, "user_id": userId, "bet_user_id": betUserId}).Select(bson.M{"_id": 1, "gold": 1}).One(&bet)
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
		sum = rst["totle"].(int)
		// sum = int(rst["totle"].(float64))
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
		sum = rst["totle"].(int)
		// revel.INFO.Println("sum = ", sum)
		// sum = int(rst["totle"].(float64))
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

// 通过投注目标玩家找到所有对他下注的记录
func FindUserBetsByBetUserId(weekId bson.ObjectId) (list []UserBet) {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_BET)
	list = []UserBet{}
	c.Find(bson.M{"bet_user_id": weekId}).All(&list)
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

// 查询对指定玩家的所有下注金额
func FindUserBetSumByBetUserId(weekId bson.ObjectId) int {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_BET)
	rst := bson.M{}
	c.Pipe([]bson.M{
		{"$match": bson.M{"bet_user_id": weekId}},
		{"$group": bson.M{
			"_id":  bson.M{"id": "$bet_user_id"},
			"gold": bson.M{"$sum": "$gold"},
		}},
	}).One(&rst)
	revel.INFO.Println("user_bet_id ")
	if val, ok := rst["gold"]; ok {
		return val.(int)
	} else {
		revel.WARN.Println("can't find bet on user_bet_id ", weekId.Hex())
		return 0
	}
}

// 发奖
func SendUserBetResultMail() {
	LoadBaseWWAWeekReward()
	WwaTypeForeach(func(typeOfWwa int) {
		SendUserBetResultMailByType(typeOfWwa)
		// delete record
		DeleteUserBetByType(typeOfWwa)
		// 参与选手发奖
		SendUserWWAWeekRewardMail(typeOfWwa)
		// 备份选手数据
		ResetScore(typeOfWwa)
	})
}

// 参与选手发奖
func SendUserWWAWeekRewardMail(typeOfWwa int) {
	list := UserWWAWeekTop20(typeOfWwa)
	if len(list) > 0 {
		firstRank := list[0]
		wwa, err := FindWWAInRedis(firstRank.ZoneId, firstRank.UserId)
		if err == nil {
			msg := fmt.Sprintf("恭喜玩家%s在巅峰之夜%s段位的对决中傲视群雄，登上至尊王座！", wwa.Name(),
				WwaTypeToName(typeOfWwa))
			BrocastNoticeToAllGameServerWithTimeInterval(msg, 3, 60)
		}
	}
	revel.INFO.Println("list = ", len(list))
	for index, week := range list {
		// 有战斗记录
		if len(week.FightedList) > 0 {
			// 发奖
			rank := index + 1
			reward := BaseWWAWeekRewardGetRewardByTypeAndRank(typeOfWwa, rank)
			if reward == "" {
				revel.WARN.Printf("can not find www reward type=%d, rank = %d \n", typeOfWwa, rank)
				continue
			}
			gs := FindGameServer(week.ZoneId)
			content := fmt.Sprintf("恭喜您在跨服竞技巅峰之夜%s段位获得第%d名的好成绩，请点击领取奖励。", WwaTypeToName(typeOfWwa), rank)
			revel.INFO.Println("content: ", content)
			url := gs.CommonRewardMail(week.UserId,
				content,
				reward)
			resp, err := GetGameServer(url)
			if err != nil {
				revel.ERROR.Println("send reward err ", err)
			} else {
				revel.INFO.Println("send reward resp ", resp)
			}
		}
	}
}

func DeleteUserBetByType(typeOfWwa int) {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_BET)
	c.RemoveAll(bson.M{"type": typeOfWwa})
}

// 按照跨服竞技类型发奖
func SendUserBetResultMailByType(typeOfWwa int) {
	revel.INFO.Println("SendUserBetResultMailByType: ", typeOfWwa)
	list := UserWWAWeekTop20(typeOfWwa)
	sys := FindSysWWAWeek()
	//test
	// CacheWWWTop3UserCache(typeOfWwa, sys, list)
	// if 1 == 1 {
	// 	return
	// }
	// 段位没有开启或者没有人, 退回所有人的元宝
	if len(list) == 0 || !sys.IsPlayoffOn[typeOfWwa] {
		sendUserBetGoldBackByWWAWeekList(typeOfWwa, list)
		return
	}
	winner := list[0]
	if winner.PlayoffScore == 0 {
		// 没有winner
		sendUserBetGoldBackByWWAWeekList(typeOfWwa, list)
		return
	}
	revel.INFO.Println("找到winner: ", winner.Id)
	// 计算总奖金
	totleBetGold := 0
	for _, week := range list {
		if len(week.FightedList) == 0 {
			// 该玩家没有参与
			SendUserBetGoldBackByWWAWeekId(typeOfWwa, &week)
		} else {
			// 参与了战斗
			curGold := FindUserBetSumByBetUserId(week.Id)
			revel.INFO.Printf("对%s一共下注了%d元宝", week.Id.Hex(), curGold)
			totleBetGold += curGold
		}
	}
	revel.INFO.Println("totleBetGold = ", totleBetGold)
	// 计算押注给冠军的总金额
	betToWinnerList := FindUserBetsByBetUserId(winner.Id)
	totleWinnerGold := 0
	for _, bet := range betToWinnerList {
		totleWinnerGold += bet.Gold
	}
	if totleWinnerGold == 0 {
		// 没有一人押中
		return
	}
	// 回报比
	rate := float64(totleBetGold) / float64(totleWinnerGold)
	rate = float64(int(rate*100)) / 100.0
	revel.INFO.Println("totleWinnerGold = ", totleWinnerGold)
	revel.INFO.Println("rate = ", rate)
	// 冠军名字
	wwa, err := FindWWAInRedis(winner.ZoneId, winner.UserId)
	if err != nil {
		revel.ERROR.Printf("can not find winner in redis zoneId = %d, userId = %d", winner.ZoneId, winner.UserId)
		return
	}
	name := wwa.Name()
	// 给押中的人发奖
	for _, bet := range betToWinnerList {
		revel.INFO.Println("SendUserBetWinnerMailByBet: ", bet.Id)
		SendUserBetWinnerMailByBet(typeOfWwa, rate, name, &bet)
	}
	CacheWWWTop3UserCache(typeOfWwa, sys, list)
}

func CacheWWWTop3UserCache(typeOfWwa int, sys *SysWWAWeek, list []UserWWAWeek) {
	// cacheData := bson.M{"_m": "", "_r": ""}
	cacheList := []map[string]interface{}{}
	for index, week := range list {
		if index >= 3 {
			break
		}
		gs := FindGameServer(week.ZoneId)
		url := gs.TopHeroInfo(week.UserId)
		data, err := GetGameServer(url)
		info := bson.M{"q": 1, "hero": 1001, "name": "未知", "rank": index + 1}
		if err == nil {
			err = json.Unmarshal(data, &info)
			if err != nil {
				revel.ERROR.Println("json unmarshal err ", err)
			} else {
				info["aid"] = week.ZoneId
				info["uid"] = week.UserId
				cacheList = append(cacheList, info)
			}
		}
	}
	// cacheData["list"] = cacheList
	// data, err := json.Marshal(cacheData)
	// if err != nil {
	// 	revel.ERROR.Println("json marshal err", err)
	// }
	// dataStr := string(data)
	// revel.INFO.Printf("cache typeOfWwa = %d, data = %s \n", typeOfWwa, dataStr)
	sys.Top3Cache[typeOfWwa] = cacheList
	UpdateSysWWAWeek(sys)
}

// 给玩家发奖
func SendUserBetWinnerMailByBet(typeOfWwa int, rate float64, name string, bet *UserBet) {
	gs := FindGameServer(bet.ZoneId)
	gold := int(float64(bet.Gold) * rate)
	url := gs.CommonRewardMail(
		bet.UserId,
		fmt.Sprintf("恭喜您在%s段位押中本赛冠军%s, 您押注金额为%d元宝, 收益倍数为:%0.2f倍。您一共获得%d元宝",
			WwaTypeToName(typeOfWwa), name, bet.Gold, rate, gold),
		fmt.Sprintf("g-%d", gold))
	GetGameServer(url)
}

// 退回所有押注给入参列表中的巅峰之夜玩家id的元宝
func sendUserBetGoldBackByWWAWeekList(typeOfWwa int, list []UserWWAWeek) {
	for _, week := range list {
		SendUserBetGoldBackByWWAWeekId(typeOfWwa, &week)
	}
}

// 将所有押注给巅峰之夜记录id的玩家的元宝退回
func SendUserBetGoldBackByWWAWeekId(typeOfWwa int, week *UserWWAWeek) {
	zoneId, userId, id := week.ZoneId, week.UserId, week.Id
	wwa, err := FindWWAInRedis(zoneId, userId)
	if err != nil {
		revel.ERROR.Println("SendUserBetGoldBackByWWAWeekId err ", err)
		return
	}
	name := wwa.Name()
	list := FindUserBetsByBetUserId(id)
	for _, bet := range list {
		sendUserBetGoldBackMail(typeOfWwa, name, &bet)
	}
}

// 押注元宝
func sendUserBetGoldBackMail(typeOfWwa int, betToName string, bet *UserBet) {
	gs := FindGameServer(bet.ZoneId)
	url := gs.CommonRewardMail(bet.UserId, fmt.Sprintf("您押注的%s段位的选手 %s 没有参赛，系统退回您的%d押注元宝",
		WwaTypeToName(typeOfWwa), betToName, bet.Gold), fmt.Sprintf("g-%d", bet.Gold))
	GetGameServer(url)
}
