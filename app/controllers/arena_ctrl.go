package controllers

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/revel/revel"
	"github.com/wangboo/wwa/app/models"
)

type ArenaCtrl struct {
	*revel.Controller
}

func (c ArenaCtrl) GroupData() revel.Result {
	return c.RenderText("")
}

func (c ArenaCtrl) TopFifty(id int) revel.Result {
	all, _ := redis.Strings(models.Redis.Do("ZRANGE", fmt.Sprintf("wwa_%d", id), 0, 49))
	return c.RenderJson(all)
}

func (c ArenaCtrl) Find() revel.Result {
	ranks := &[]models.Rank{}
	models.DB.Where("name=?", "wangbo").Find(ranks)
	fmt.Println("len = ", len(*ranks))
	return c.RenderJson(ranks)
}
