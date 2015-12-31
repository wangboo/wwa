package controllers

import (
	"github.com/revel/revel"
	"github.com/wangboo/wwa/app/models"
	"net/url"
)

type CronTabCtrl struct {
	*revel.Controller
}

func (c *CronTabCtrl) AddScrollMsg(exp, name, message, zoneIds string) revel.Result {
	exp, _ = url.QueryUnescape(exp)
	name, _ = url.QueryUnescape(name)
	message, _ = url.QueryUnescape(message)
	zoneIds, _ = url.QueryUnescape(zoneIds)
	revel.INFO.Printf("exp = %s, name = %s, message = %s, zoneIds = %s \n", exp, name, message, zoneIds)
	err := models.AddTimerScollMsg(exp, name, message, zoneIds)
	if err != nil {
		return c.RenderJson(FailWithError(err))
	}
	return c.RenderJson(Succ())
}

func (c *CronTabCtrl) List() revel.Result {
	return c.RenderJson(Succ("list", models.All()))
}

// 取消任务
func (c *CronTabCtrl) Delete(id int) revel.Result {
	models.DeleteCronTab(id)
	return c.List()
}
