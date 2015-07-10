package models

import (
	"encoding/json"
	"github.com/revel/revel"
	"io/ioutil"
	"os"
	"path/filepath"
)

type BaseInv struct {
	Type      string `json:"type"`
	SubType   string `json:"sub_type"`
	Id        int    `json:"id"`
	Size      int    `json:"size"`
	Condition int    `json:"condition"`
	Reward    string `json:"reward"`
	Desc      string `json:"desc"`
}

var (
	BaseMainInvList  []BaseInv
	BaseDailyInvList []BaseInv
)

func InitBaseInvitation() {
	root, ok := revel.Config.String("root")
	if !ok {
		panic("app.conf中找不到root配置")
	}
	configPath := filepath.Join(root, "conf", "invitation.json")
	configFile, err := os.Open(configPath)
	if err != nil {
		panic(err)
	}
	data, _ := ioutil.ReadAll(configFile)
	BaseInvList := []BaseInv{}
	BaseMainInvList = []BaseInv{}
	BaseDailyInvList = []BaseInv{}
	json.Unmarshal(data, &BaseInvList)
	for _, v := range BaseInvList {
		if v.Type == "MAIN" {
			BaseMainInvList = append(BaseMainInvList, v)
		} else if v.Type == "DAILY" {
			BaseDailyInvList = append(BaseDailyInvList, v)
		} else {
			revel.ERROR.Println("未知的邀请奖励：", v)
		}
	}
	revel.INFO.Println("邀请主线任务奖励:", BaseMainInvList)
	revel.INFO.Println("邀请每日任务奖励:", BaseDailyInvList)
}

func FindBaseDailyTaskId(taskId int) *BaseInv {
	for _, o := range BaseDailyInvList {
		if o.Id == taskId {
			return &o
		}
	}
	return nil
}

func FindBaseMaidTaskId(taskId int) *BaseInv {
	for _, o := range BaseMainInvList {
		if o.Id == taskId {
			return &o
		}
	}
	return nil
}
