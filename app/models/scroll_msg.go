package models

import (
	"github.com/revel/revel"
	// "labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strconv"
	"strings"
)

const (
	COL_CRON_TAB_SCROLL_MSG = "cron_tab_scroll_msg"
)

func InitScrollMsg() {
	SetEvent(0, cronTriggerScollMsg)
}

type CronTabScrollMsg struct {
	Id        bson.ObjectId `bson:"_id"`
	CronTabId bson.ObjectId `bson:"cron_tab_id"`
	Prefix    string        `bson:"prefix"`
	Msg       string        `bson:"msg"`
	ZoneIds   string        `bson:"zone_ids"` // 1,2,3,4 id集合，或者all代表所有
}

type CronTabScrollMsgCombin struct {
	Msg  CronTabScrollMsg
	Cron CronTab
}

func AddTimerScollMsg(exp, prefix, msg, zoneIds string) error {
	s := Session()
	defer s.Close()
	c := Col(s, COL_CRON_TAB_SCROLL_MSG)
	cron, _, err := AddTimer(exp, "", 0)
	if err != nil {
		return err
	}
	msgItem := CronTabScrollMsg{
		Id:        bson.NewObjectId(),
		CronTabId: cron.Id,
		Prefix:    prefix,
		Msg:       msg,
		ZoneIds:   zoneIds,
	}
	c.Insert(&msgItem)
	return nil
}

// 触发定时任务
func cronTriggerScollMsg(item *CronTab) {
	s := Session()
	defer s.Close()
	c := Col(s, COL_CRON_TAB_SCROLL_MSG)
	msg := CronTabScrollMsg{}
	if err := c.Find(bson.M{"cron_tab_id": item.Id}).One(&msg); err == nil {
		if msg.ZoneIds == "all" {
			// 所有服务器
			BrocastNoticeToAllGameServer(msg.Msg)
		} else {
			for _, zoneIdStr := range strings.Split(msg.ZoneIds, ",") {
				zoneId, err := strconv.Atoi(zoneIdStr)
				if err == nil {
					if gs := FindGameServer(zoneId); gs != nil {
						url := gs.NoticeAdvanceUrl(msg.Prefix, msg.Msg)
						revel.INFO.Println("scroll_msg: ", url)
						GetGameServer(url)
					}
				} else {
					revel.ERROR.Println("ZoneIds错误：", msg.ZoneIds)
				}
			}
		}
	}
}

func All() (all []*CronTabScrollMsgCombin) {
	s := Session()
	defer s.Close()
	cc := Col(s, COL_CRON_TAB)
	cs := Col(s, COL_CRON_TAB_SCROLL_MSG)
	cronAll := []CronTab{}
	msgAll := []CronTabScrollMsg{}
	cc.Find(nil).All(&cronAll)
	cs.Find(nil).All(&msgAll)

	// all := []*CronTabScrollMsgCombin{}
	for _, cron := range cronAll {
		for _, msg := range msgAll {
			if msg.CronTabId.Hex() == cron.Id.Hex() {
				item := &CronTabScrollMsgCombin{Cron: cron, Msg: msg}
				all = append(all, item)
				break
			}
		}
	}
	return
}
