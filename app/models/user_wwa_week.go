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
	IsFinish     bool            `bson:"is_finish,omitempty"`   // 是否挑战完毕
	PlayoffScore int             `bson:"playoff_score"`         // 季后赛积分
	CreatedAt    time.Time       `bson:"created_at"`            // 创建时间
	IsSend       bool            `bson:"is_send"`               // 是否发放奖励
	InTop20      bool            `bson:"in_top20"`              // 是否进入季后赛前20
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
	COL_SYS_WWA_WEEK         = "sys_wwa_weeks"         // 系统巅峰之夜状态表
	// 巅峰之夜每个段位参与挑战人数
	WWA_WEEK_RANK_SIZE_LIMIT = 20
	// 巅峰之夜每个段位开赛最小人数
	WWA_WEEK_PLAYOFF_ON_MIN_SIZE = 3 // 正式环境为 5
)

// 系统巅峰竞技状态
type SysWWAWeek struct {
	Id          bson.ObjectId              `bson:"_id"`           // 主键
	State       int                        `bson:"state"`         // 巅峰竞技状态
	IsPlayoffOn []bool                     `bson:"is_playoff_on"` // 巅峰竞技是否满足开赛条件
	Top3Cache   [][]map[string]interface{} `bson:"top3_cache"`    // 每个段位前3
	SysBets     []int                      `bson:sys_bets`        // 系统押注补发元宝
}

const (
	TYPE_WWW_VIEW_NORMAL      = 0 // 周一到周六显示普通挑战积分（周一至周六，周六11点结算常规赛积分）
	TYPE_WWW_VIEW_WAIT        = 1 // 准备界面11点到12点，不能点击
	TYPE_WWW_VIEW_BET_ENABLE  = 2 // 内部状态，下主界面，可以下注
	TYPE_WWW_VIEW_BET_DISABLE = 3 // 内部状态，下主界面，不能下注
	TYPE_WWW_VIEW_FIGHT_IN    = 4 // 周日挑战界面（周日20点-21点）参与人员
	TYPE_WWW_VIEW_FIGHT_OUT   = 5 // 周日挑战界面（周日20点-21点）非参与人员
	TYPE_WWW_VIEW_OVER        = 6 // 周日挑战结果界面（周日21点-周日0点）
	TYPE_WWW_VIEW_PREPARE     = 7 // 周六23点到24点巅峰之夜准备中

)

func CreateUserWeekWWAFightResult(fightId bson.ObjectId, rst, score int) *FightResult {
	return &FightResult{
		UserId:   fightId,
		FightRst: rst,
		Score:    score,
	}
}

// 查找或者创建玩家巅峰之夜数据
func FindOrCreateWWAWeekByZoneIdAndUserId(zoneId, userId int) (week *UserWWAWeek, err error) {
	return findWWAWeekByZoneIdAndUserIdWithCreateFlag(zoneId, userId, true)
}

// 通过zoneId和userId查询玩家周巅峰之夜数据
func FindWWAWeekByZoneIdAndUserId(zoneId, userId int) (week *UserWWAWeek, err error) {
	return findWWAWeekByZoneIdAndUserIdWithCreateFlag(zoneId, userId, false)
}

// 查找或者创建玩家巅峰之夜数据
func findWWAWeekByZoneIdAndUserIdWithCreateFlag(zoneId, userId int, create bool) (week *UserWWAWeek, err error) {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_WWA_WEEK)
	week = &UserWWAWeek{}
	err = c.Find(bson.M{"zone_id": zoneId, "user_id": userId}).One(week)
	if err != nil {
		if err == mgo.ErrNotFound && create {
			wwa, redisErr := FindWWAInRedis(zoneId, userId)
			if redisErr != nil {
				revel.ERROR.Println("redis error:", redisErr)
				return nil, redisErr
			}
			week.Id = bson.NewObjectId()
			week.ZoneId = zoneId
			week.UserId = userId
			week.Pow = wwa.Pow()
			week.InTop20 = false
			if IsWeekend() {
				week.Score = 0
			} else {
				week.Score = wwa.Score()
			}
			week.Type = wwa.Type()
			week.CreatedAt = time.Now()
			week.FightedList = []FightResult{}
			week.WaitList = []bson.ObjectId{}
			week.IsFinish = false
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
	if IsWeekend() {
		// 星期天不计入季前赛积分
		return false
	}
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_WWA_WEEK)
	week, err := FindOrCreateWWAWeekByZoneIdAndUserId(zoneId, userId)
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
	c.Find(bson.M{"type": typeOfWwa, "in_top20": true}).Sort("-playoff_score", "-pow").Limit(WWA_WEEK_RANK_SIZE_LIMIT).All(&list)
	return
}

func UserWWAWeekTop20ByScore(typeOfWwa int) (list []UserWWAWeek) {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_WWA_WEEK)
	list = []UserWWAWeek{}
	c.Find(bson.M{"type": typeOfWwa, "score": bson.M{"$gt": 0}}).Sort("-score", "-pow").Limit(WWA_WEEK_RANK_SIZE_LIMIT).All(&list)
	return
}

// 找到第一名
func FindUserWWAWeekFirstRank(typeOfWwa int) (week *UserWWAWeek, err error) {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_WWA_WEEK)
	list := []UserWWAWeek{}
	err = c.Find(bson.M{"type": typeOfWwa}).Sort("-playoff_score", "-pow").Limit(1).All(&list)
	if len(list) == 0 {
		if err != nil {
			return nil, err
		} else {
			// 找不到记录
			return week, mgo.ErrNotFound
		}
	}
	return &list[0], err
}

// 巅峰之夜20名单
func UserWWAWeekPlayoff(typeOfWwa int) (list []UserWWAWeek) {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_WWA_WEEK)
	list = []UserWWAWeek{}
	c.Find(nil).Sort("-playoff_score", "-pow").Limit(WWA_WEEK_RANK_SIZE_LIMIT).All(&list)
	return
}

// 切换到季后赛模式，删除20名以后的人员(指定类型)
func UserWWAWeekSwitch2PlayoffByType(typeOfWwa int) {
	list := UserWWAWeekTop20ByScore(typeOfWwa)
	sizeOfList := len(list)
	sys := FindSysWWAWeek()
	if sizeOfList < WWA_WEEK_PLAYOFF_ON_MIN_SIZE {
		// 不满足开赛条件
		sys.IsPlayoffOn[typeOfWwa] = false
		UpdateSysWWAWeek(sys)
		return
	} else {
		sys.IsPlayoffOn[typeOfWwa] = true
		UpdateSysWWAWeek(sys)
	}
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_USER_WWA_WEEK)
	c.RemoveAll(bson.M{"type": typeOfWwa})
	// 初始化挑战玩家数据
	for i := 0; i < sizeOfList; i++ {
		week := list[i]
		AllOtherIds := []bson.ObjectId{}
		for _, item := range list {
			if week.Id != item.Id {
				AllOtherIds = append(AllOtherIds, item.Id)
			}
		}
		if len(week.FightedList) > 0 {
			week.FightedList = []FightResult{}
		}
		week.FightingId = AllOtherIds[0]
		week.WaitList = AllOtherIds[1:]
		week.InTop20 = true
		c.Insert(week)
	}
}

// 切换到季后赛模式，删除20名以后的人员
func UserWWAWeekSwitch2Playoff() {
	WwaTypeForeach(func(typeOfWwa int) {
		UserWWAWeekSwitch2PlayoffByType(typeOfWwa)
	})
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

// 获取系统巅峰竞技状态
func FindSysWWAWeek() *SysWWAWeek {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_SYS_WWA_WEEK)
	week := &SysWWAWeek{}
	err := c.Find(nil).One(week)
	if err != nil {
		if err == mgo.ErrNotFound {
			// week.State = TYPE_WWW_VIEW_NORMAL
			week.Id = bson.NewObjectId()
			week.IsPlayoffOn = []bool{false, false, false, false}
			week.Top3Cache = [][]map[string]interface{}{}
			week.SysBets = []int{0, 0, 0, 0}
			c.Insert(week)
		}
	}
	if len(week.Top3Cache) == 0 {
		cacheSample := []map[string]interface{}{}
		week.Top3Cache = [][]map[string]interface{}{}
		week.Top3Cache = append(week.Top3Cache, cacheSample)
		week.Top3Cache = append(week.Top3Cache, cacheSample)
		week.Top3Cache = append(week.Top3Cache, cacheSample)
		week.Top3Cache = append(week.Top3Cache, cacheSample)
	}
	if len(week.SysBets) == 0 {
		week.SysBets = []int{0, 0, 0, 0}
	}
	return week
}

// 更新系统巅峰竞技状态
func UpdateSysWWAWeek(sys *SysWWAWeek) {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_SYS_WWA_WEEK)
	c.UpdateId(sys.Id, sys)
}

func (u *SysWWAWeek) UpdateGold() {
	s := Session()
	defer s.Close()
	c := s.DB(DB_NAME).C(COL_SYS_WWA_WEEK)
	c.UpdateId(u.Id, bson.M{"$set": bson.M{"sys_bets": u.SysBets}})
}

func IsWeekend() bool {
	return time.Now().Weekday() == 0 || true
}

func GetSysWWAWeekState() int {
	now := time.Now()
	weekday := now.Weekday()
	hour := now.Hour()
	// test
	weekday, hour = 2, 10
	// 结果
	// switch weekday {
	// case 0:
	if weekday == 0 {
		// 周天
		if hour < 10 {
			return TYPE_WWW_VIEW_BET_DISABLE
		} else if hour < 19 {
			// 下注界面
			return TYPE_WWW_VIEW_BET_ENABLE
		} else if hour < 21 {
			// 挑战界面
			return TYPE_WWW_VIEW_FIGHT_IN
		} else {
			return TYPE_WWW_VIEW_OVER
		}
	} else if weekday == 1 {
		if hour < 12 {
			// 周一1点前
			return TYPE_WWW_VIEW_OVER
		} else {
			return TYPE_WWW_VIEW_NORMAL
		}
	} else if weekday == 6 {
		if hour >= 23 {
			// 周六 提示：巅峰之夜准备中，请在周日0点以后查看
			return TYPE_WWW_VIEW_PREPARE
		} else {
			// 显示
			return TYPE_WWW_VIEW_NORMAL
		}
	} else {
		// 周一到周五
		return TYPE_WWW_VIEW_NORMAL
	}
}
