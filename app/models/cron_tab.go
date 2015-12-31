package models

import (
	"github.com/revel/revel"
	"github.com/wangboo/cron"
	// "labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

const (
	COL_CRON_TAB           = "cron_tab"
	CRON_TAB_STATE_NEW     = 0 // 新的，还没运行
	CRON_TAB_STATE_RUN     = 1 // 正在运行中
	CRON_TAB_STATE_EXPIRED = 2 // 已经过期
)

var (
	cronJob *cron.Cron
	jobMap  map[int]func(*CronTab)
)

type CronTab struct {
	Id     bson.ObjectId `bson:"_id"`
	Exp    string        `bson:"exp"`
	Type   int           `bson:"type"`    //0滚动消息
	State  int           `bson:"state"`   // 当前状态
	Times  int           `bson:"times"`   // 已经运行次数
	Desc   string        `bson:"desc"`    // 秒数
	CronId int           `bson:"cron_id"` // 任务id
}

// 初始化，加载所有配置的定时任务
func InitCronTab() {
	jobMap = map[int]func(*CronTab){}
	cronJob = cron.New()
	s := Session()
	defer s.Close()
	c := Col(s, COL_CRON_TAB)
	all := []CronTab{}
	if err := c.Find(nil).All(&all); err != nil {
		panic(err)
	}
	for _, item := range all {
		revel.INFO.Println("add cron job", item.Exp)
		id, _ := cronJob.AddFunc(item.Exp, func() {
			cronTrigger(&item)
		})
		item.CronId = id
		item.Update()
	}
	cronJob.Start()
}

// 这是不同配置任务的触发函数
func SetEvent(ctype int, f func(*CronTab)) {
	jobMap[ctype] = f
}

func AddTimer(exp, desc string, ctype int) (item *CronTab, id int, err error) {
	item = &CronTab{
		Id:    bson.NewObjectId(),
		Exp:   exp,
		Type:  ctype,
		State: CRON_TAB_STATE_NEW,
		Times: 0,
		Desc:  desc,
	}
	id, err = cronJob.AddFunc(exp, func() { cronTrigger(item) })
	if err != nil {
		return
	}
	item.CronId = id
	s := Session()
	defer s.Close()
	c := Col(s, COL_CRON_TAB)
	c.Insert(item)

	return
}

// 更新
func (c *CronTab) Update() {
	s := Session()
	defer s.Close()
	col := Col(s, COL_CRON_TAB)
	col.UpdateId(c.Id, c)
}

// 删除指定任务
func DeleteCronTab(cronId int) {
	s := Session()
	defer s.Close()
	cc := Col(s, COL_CRON_TAB)
	cs := Col(s, COL_CRON_TAB_SCROLL_MSG)
	cronJob.Remove(cronId)
	item := CronTab{}
	cc.Find(bson.M{"cron_id": cronId}).One(&item)
	cc.Remove(bson.M{"cron_id": cronId})
	cs.Remove(bson.M{"cron_tab_id": item.Id})
}

func cronTrigger(item *CronTab) {
	if f, ok := jobMap[item.Type]; ok {
		item.Times += 1
		item.Update()
		f(item)
	} else {
		revel.ERROR.Println("unknown cronTrigger type", item.Type)
	}
}
