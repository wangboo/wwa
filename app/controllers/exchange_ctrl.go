package controllers

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/revel/revel"
	"github.com/wangboo/wwa/app/models"
	"log"
)

type ExchangeCtrl struct {
	*revel.Controller
}

// 重新载入配置文件
func (c ExchangeCtrl) ReloadConfig() revel.Result {
	err := models.ReloadExchangeConfig()
	if err != nil {
		return c.RenderText("加载失败：%s", err.Error())
	}
	return c.RenderJson(models.BaseExchangeList)
}

// 重新生成每日兑换
func (c ExchangeCtrl) ResetDailyExchange() revel.Result {
	err := models.ResetDailyExchange()
	if err != nil {
		return c.RenderText("重置失败：%s", err.Error())
	}
	return c.RenderJson(models.TodayExchangeList)
}

// 查询每日兑换内容
func (c ExchangeCtrl) DailyExchange() revel.Result {
	if len(models.TodayExchangeList) > 0 {
		return c.RenderJson(models.TodayExchangeList)
	}
	cli := models.RedisPool.Get()
	defer cli.Close()
	data, err := redis.String(cli.Do("GET", models.TodayExchangeKey))
	if err != nil || len(data) == 0 {
		models.ReloadExchangeConfig()
		return c.ResetDailyExchange()
	}
	log.Printf("get TodayExchangeList from redis %s\n", data)
	json.Unmarshal([]byte(data), &models.TodayExchangeList)
	return c.RenderJson(models.BaseExchangeList)
}
