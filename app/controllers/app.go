package controllers

import (
	"github.com/revel/revel"
	"github.com/wangboo/wwa/app/jobs"
	"strings"
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
