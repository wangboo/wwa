package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/revel/revel"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"math/rand"
	"net/http"
	"strconv"
)

var (
	randomCodes = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q",
		"r", "s", "t", "u", "v", "w", "x", "y", "z", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
		"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
)

type UserInv struct {
	Id            bson.ObjectId     `bson:"_id"`
	Username      string            `bson:"username"`
	ZoneId        int               `bson:"zone_id"`
	Code          string            `bson:"code"`
	Focus         bson.ObjectId     `bson:"focus,omitempty"` // 我使用别人激活码的用户id
	TodayInvTimes int               `bson:"today_inv_times"` // 今日推广次数
	DailyTask     []*UserTaskStatus `bson:"daily_task"`      // 每日任务完成集合
	MainTask      []*UserTaskStatus `bson:"main_task"`       // 主线任务
}

// 任务进度
type UserTaskStatus struct {
	TaskId        int             `bson:"task_id"`  // 任务id
	CompletionIds []bson.ObjectId `bson:"user_ids"` // 完成的玩家
	Complete      bool            `bson:"complete"` // 是否完成
	Got           bool            `bson:"got"`      // 是否领取奖励
}

// 玩家任务信息
type UserTaskInfo struct {
	Level int `json:"level"`
	Vip   int `json:"vip"`
}

// 前台显示信息
type UserInvShowInfo struct {
	Zone  string
	Name  string `json:"name"`
	Level int    `json:"level"`
	Vip   int    `json:"vip"`
	Pow   int    `json:"pow"`
	Frame int    `json:"frame"`
	Photo int    `json:"photo"`
}

func NewUserInv(c *mgo.Collection, username string, zoneId int) (user *UserInv, err error) {
	user, err = FindUserInv(c, username, zoneId)
	if err == nil {
		return user, fmt.Errorf("%s已经存在", username)
	} else if err != mgo.ErrNotFound {
		return user, fmt.Errorf("%s已经存在(err: %s)", username, err.Error())
	}
	newCode, err := newUserInvCode(c, zoneId)
	if err != nil {
		return nil, err
	}
	user.Id = bson.NewObjectId()
	user.Username = username
	user.ZoneId = zoneId
	user.Code = newCode
	user.DailyTask = []*UserTaskStatus{}
	user.MainTask = []*UserTaskStatus{}
	for _, bd := range BaseDailyInvList {
		userDaily := &UserTaskStatus{bd.Id, []bson.ObjectId{}, false, false}
		user.DailyTask = append(user.DailyTask, userDaily)
	}
	for _, bm := range BaseMainInvList {
		userMain := &UserTaskStatus{bm.Id, []bson.ObjectId{}, false, false}
		user.MainTask = append(user.MainTask, userMain)
	}
	err = c.Insert(user)
	return user, err
}

func FindUserInv(c *mgo.Collection, username string, zoneId int) (user *UserInv, err error) {
	user = &UserInv{}
	err = c.Find(bson.M{"username": username, "zone_id": zoneId}).One(user)
	return
}

func newUserInvCode(c *mgo.Collection, zoneId int) (string, error) {
	buffer := bytes.NewBufferString(strconv.Itoa(zoneId))
	lenOfCodes := len(randomCodes)
	for i := 0; i < 7; i++ {
		randNum := rand.Intn(lenOfCodes)
		buffer.WriteString(randomCodes[randNum])
	}
	code := buffer.String()
	revel.INFO.Println("generate code :", code)
	if _, err := FindUserInvByCode(c, code); err != nil {
		if err == mgo.ErrNotFound {
			return code, nil
		}
		revel.ERROR.Println("newUserInvCode err :", err)
		return "", err
	}
	return newUserInvCode(c, zoneId)
}

// 根据激活码查询玩家记录
func FindUserInvByCode(c *mgo.Collection, code string) (user *UserInv, err error) {
	user = &UserInv{}
	err = c.Find(bson.M{"code": code}).One(user)
	return user, err
}

// 通过关注人查询被关注人信息
func FindUserInvByFocusId(c *mgo.Collection, id bson.ObjectId) (user *UserInv, err error) {
	err = c.Find(bson.M{"focus": id}).One(user)
	return user, err
}

// 查询所有玩家信息
func FindUserTaskInfos(users []UserInv) []*UserTaskInfo {
	all := []*UserTaskInfo{}
	for _, u := range users {
		gs := FindGameServer(u.ZoneId)
		url := gs.UserLevelAndVipUrl(u.Username)
		revel.INFO.Println("url = ", url)
		resp, err := http.Get(url)
		info := &UserTaskInfo{}
		if err != nil {
			revel.ERROR.Println("FindUserTaskInfos http.Get err :", err.Error())
			continue
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			revel.ERROR.Println("FindUserTaskInfos ioutil.ReadAll err :", err.Error())
			continue
		}
		revel.INFO.Printf("resp = %s \n", data)
		err = json.Unmarshal(data, info)
		revel.INFO.Println(info)
		if err != nil {
			revel.ERROR.Println("FindUserTaskInfos json Unmarshal err :", err.Error())
		}
		all = append(all, info)
	}
	return all
}
