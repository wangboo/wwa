package app

import (
	// "fmt"
	"github.com/revel/modules/jobs/app/jobs"
	"github.com/revel/revel"
	// "github.com/wangboo/wwa/app/controllers"
	"github.com/wangboo/wwa/app/jobs"
	"github.com/wangboo/wwa/app/models"
	"log"
	"reflect"
	"strings"
	// "unsafe"
)

func init() {
	// Filters is the default set of global filters.
	revel.Filters = []revel.Filter{
		revel.PanicFilter,             // Recover from panics and display an error page instead.
		revel.RouterFilter,            // Use the routing table to select the right Action
		revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
		revel.ParamsFilter,            // Parse parameters into Controller.Params.
		revel.SessionFilter,           // Restore and write the session cookie.
		revel.FlashFilter,             // Restore and write the flash cookie.
		revel.ValidationFilter,        // Restore kept validation errors and save new ones from cookie.
		revel.I18nFilter,              // Resolve the requested language
		//HeaderFilter,                  // Add some security based headers
		revel.InterceptorFilter, // Run interceptors around the action.
		revel.CompressFilter,    // Compress the result.
		WhiteIPFilter,
		MgoSessionFilter,
		revel.ActionInvoker, // Invoke the action.
	}
	// register startup functions with OnAppStart
	revel.OnAppStart(models.InitGameServerConfig)
	revel.OnAppStart(models.InitBaseInvitation)
	revel.OnAppStart(models.InitDatabase)
	revel.OnAppStart(models.InitRedis)
	revel.OnAppStart(func() {
		// 日终竞技场排名清空
		jobs.Schedule("0 15 4 * * ?", &mjob.RankDataJob{})
		// 日终兑换清空
		jobs.Schedule("0 0 0 * * ?", &mjob.ExchangeJob{})
		// 跨服竞技场奖励
		jobs.Schedule("45 22 0 * * ?", &mjob.DayEndRewardJob{})
		// jobs.Now(&mjob.RankDataJob{})
		jobs.Schedule("0 0 0 * * ?", &mjob.UserInvDailyJobReset{})
		// 周六结算
		jobs.Schedule("0 10 23 * * 6", &mjob.WWAWeekFightBeginJob{})
		// 周天21点
		jobs.Schedule("10 0 21 * * 0", &mjob.WWWWeekFightEndJob{})
		// 巅峰之夜数据清空
		jobs.Schedule("1 1 0 * * 1", &mjob.WWWWeekCleanJob{})
	})
}

var WhiteIPFilter = func(c *revel.Controller, fc []revel.Filter) {
	enable := revel.Config.BoolDefault("GameServer.whiteIP", false)
	if !enable {
		fc[0](c, fc[1:])
		return
	}
	uri := c.Request.RequestURI
	if strings.HasPrefix(uri, "/wwa/") || strings.HasPrefix(uri, "/inv/") {
		remoteIP := c.Request.Header.Get("X-Forwarded-For")
		log.Printf("remote : %s\n", remoteIP)
		for _, s := range models.GameServerList {
			if s.Ip == remoteIP {
				fc[0](c, fc[1:])
				return
			}
		}
		c.Result = c.RenderText("Unauthorized")
	} else {
		fc[0](c, fc[1:])
	}
}

// mgo Session 注入
var MgoSessionFilter = func(c *revel.Controller, fc []revel.Filter) {
	appCtrl := c.AppController
	typeOfC := reflect.TypeOf(appCtrl).Elem()
	_, ok := typeOfC.FieldByName("MSession")
	if !ok {
		fc[0](c, fc[1:])
		return
	}
	valueOfC := reflect.ValueOf(appCtrl).Elem()
	// 注入session
	newSession := models.Session()
	defer newSession.Close()
	valueOfSession := reflect.ValueOf(newSession)
	valoeOfElem := valueOfC.FieldByName("MSession")
	valoeOfElem.Set(valueOfSession)
	fc[0](c, fc[1:])
}
