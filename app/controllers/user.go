package controllers

import (
	"github.com/revel/revel"
	"github.com/wangboo/wwa/app/models"
	"net/http"
)

type User struct {
	*revel.Controller
}

// 转发到其他游戏服务器
func (c *User) ChangeServer(name string, areaId int) revel.Result {
	for _, gs := range models.GameServerList {
		if gs.ZoneId == areaId {
			continue
		}
		url := gs.ChangeServerUrl()
		go http.Get(url)
	}
	return c.RenderText("ok")
}
