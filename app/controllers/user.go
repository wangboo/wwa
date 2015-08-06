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
func (c *User) ChangeServer(token string, areaId int) revel.Result {
	for _, gs := range models.GameServerList {
		if gs.ZoneId == areaId {
			continue
		}
		go func() {
			url := gs.ChangeServerUrl(token)
			resp, err := http.Get(url)
			if err != nil {
				revel.ERROR.Println("get error", err)
				return
			}
			resp.Body.Close()
		}()

	}
	return c.RenderText("ok")
}

func (c *User) InitGameServerConfig() revel.Result {
	models.InitGameServerConfig()
	return c.RenderJson(models.GameServerList)
}
