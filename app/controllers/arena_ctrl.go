package controllers

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	// "github.com/revel/modules/jobs/app/jobs"
	"github.com/revel/revel"
	"github.com/wangboo/wwa/app/jobs"
	"github.com/wangboo/wwa/app/models"
	"regexp"
	"strconv"
)

var (
	GetScoreRegex = regexp.MustCompile(`^(\d+,)(\d+)(.*(\d+))$`)
)

type ArenaCtrl struct {
	*revel.Controller
}

func (c ArenaCtrl) GroupData(a, u int) revel.Result {
	gi, err := models.Redis.Do("GET", models.GroupDataKey(u, a))
	if err != nil {
		c.RenderText("redis err %v", err)
	}
	if gi != nil {
		return c.RenderText("%s", gi)
	}
	str := models.GetGroupDataFromGameServer(a, u)
	if len(str) == 0 {
		return c.RenderText("")
	}
	sec := revel.Config.IntDefault("group_data_cache_time", 60)
	models.Redis.Do("SETEX", models.GroupDataKey(u, a), sec, str)
	return c.RenderText(str)
}

func (c ArenaCtrl) TopFifty(id int) revel.Result {
	allSimplyKeys, _ := redis.Strings(models.Redis.Do("ZRANGE", fmt.Sprintf("wwa_%d", id), 0, 49))
	args := redis.Args{}
	args = args.Add("zone_user")
	for _, k := range allSimplyKeys {
		args = args.Add(k)
	}
	fmt.Printf("args = %v \n", args)
	all, err := redis.Strings(models.Redis.Do("HMGET", args...))
	if err != nil {
		return c.RenderText("redis err", err.Error())
	}
	return c.RenderJson(all)
}

// 获得积分 或者 消费积分
func (c ArenaCtrl) GetScore(s, u, a int) revel.Result {
	c.Validation.Required(s).Message("您必须提供 s:获得积分")
	c.Validation.Required(u).Message("您必须提供 u:用户编号")
	c.Validation.Required(a).Message("您必须提供 a:服务器编号")
	c.Validation.Min(a, 0).Message("a:服务器编号必须大于0")
	c.Validation.Min(s, 0).Message("s:获取积分必须大于0")
	if c.Validation.HasErrors() {
		return RenderValidationFail(c.Controller)
	}
	detail, err := models.Redis.Do("HGET", "zone_user", models.ToSimpleKey(a, u))
	if err != nil {
		return c.RenderText("redis err %v", err)
	}
	if detail == nil {
		return c.RenderText("user not exists")
	}
	fmt.Printf("%s\n", detail)
	rst := GetScoreRegex.FindStringSubmatch(fmt.Sprintf("%s", detail))
	score, _ := strconv.Atoi(rst[2])
	newScore := score + s
	newStr := fmt.Sprintf("%s%d%s", rst[1], newScore, rst[3])
	fmt.Println("save newStr : ", newStr)
	simpleKey := models.ToSimpleKey(a, u)
	// 更新缓存数据
	models.Redis.Do("HSET", "zone_user", simpleKey, newStr)
	models.Redis.Do("ZADD", fmt.Sprintf("wwa_%s", rst[4]), fmt.Sprintf("%d", score-s), simpleKey)
	return c.RenderText("%s", newStr)
}

// 队伍信息
func (c ArenaCtrl) GroupInfo(u, a int) revel.Result {
	gi, err := models.Redis.Do("GET", models.GroupInfoKey(u, a))
	if err != nil {
		c.RenderText("redis err %v", err)
	}
	if gi != nil {
		return c.RenderText("%s", gi)
	}
	str := models.GetGroupInfoFromGameServer(a, u)
	if len(str) == 0 {
		return c.RenderText("")
	}
	sec := revel.Config.IntDefault("group_info_cache_time", 60)
	models.Redis.Do("SETEX", models.GroupInfoKey(u, a), sec, str)
	return c.RenderText(str)
}

// 基本信息
func (c ArenaCtrl) BaseInfo(u, a int) revel.Result {
	detail, err := models.Redis.Do("HGET", "zone_user", models.ToSimpleKey(a, u))
	if err != nil {
		return c.RenderText("redis error", err.Error())
	}
	return c.RenderText("%s", detail)
}

func (c ArenaCtrl) RankType(u, a int) revel.Result {
	c.Validation.Required(u).Message("您必须提供 u:用户编号")
	c.Validation.Required(a).Message("您必须提供 a:服务器编号")
	if c.Validation.HasErrors() {
		return RenderValidationFail(c.Controller)
	}
	detail, err := models.Redis.Do("HGET", "zone_user", models.ToSimpleKey(a, u))
	if err != nil {
		return c.RenderText("redis error", err.Error())
	}
	rst := GetScoreRegex.FindStringSubmatch(fmt.Sprintf("%s", detail))
	return c.RenderText("%s", rst[4])
}

func (c ArenaCtrl) ResetRank(pwd string) revel.Result {
	if pwd != revel.Config.StringDefault("password", "w231520") {
		c.RenderText("fail password error")
	}
	(&mjob.RankDataJob{}).Run()
	return c.RenderText("success")
}
