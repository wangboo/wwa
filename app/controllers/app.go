package controllers

import (
	"github.com/revel/revel"
	"github.com/wangboo/wwa/app/jobs"
	"github.com/wangboo/wwa/app/models"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	return c.Render()
}

func RenderValidationFail(c *revel.Controller) revel.Result {
	errorMessage := make([]string, len(c.Validation.Errors))
	for i := 0; i < len(c.Validation.Errors); i++ {
		errorMessage[i] = c.Validation.Errors[i].Message
	}
	return c.RenderText(strings.Join(errorMessage, ","))
}

func (c App) Test() revel.Result {
	job := &mjob.DayEndRewardJob{}
	job.Run()
	return c.RenderText("ok")
}

//
func (c App) Payment(begindt string, enddt string) revel.Result {
	_, err := time.Parse("2006-01-02", begindt)
	if err != nil {
		return c.RenderText("begindt format error(yyyy-MM-dd): %s", err.Error())
	}
	_, err = time.Parse("2006-01-02", enddt)
	if err != nil {
		return c.RenderText("enddt format error(yyyy-MM-dd): %s", err.Error())
	}
	value := url.Values{}
	value.Add("begindt", begindt)
	value.Add("enddt", enddt)
	length := len(models.GameServerList)
	ch := make(chan int, length)
	for _, gs := range models.GameServerList {
		go getPaymentFromGameServer(ch, gs, value)
	}
	sum := 0
	for i := 0; i < length; i++ {
		select {
		case payment := <-ch:
			sum += payment
		}
	}
	return c.RenderText("%d", sum)
}

func getPaymentFromGameServer(ch chan int, gs models.GameServerConfig, data url.Values) {
	reqUrl := gs.Payment()
	rstBytes, err := models.PostFormGameServer(reqUrl, data)
	if err != nil {
		revel.ERROR.Println("服务器 %s 响应错误 %s ", reqUrl, err.Error())
		ch <- 0
		return
	}
	payment, err := strconv.Atoi(string(rstBytes))
	if err != nil {
		revel.ERROR.Printf("金额转换错误 url = %s, payment = %s \n", reqUrl, rstBytes)
		ch <- 0
		return
	}
	ch <- payment
}
