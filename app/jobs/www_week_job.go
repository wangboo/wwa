package mjob

import (
	// "fmt"
	// "github.com/revel/revel"
	"github.com/wangboo/wwa/app/models"
	// "time"
)

// 挑战前初始化定时器
type WWAWeekFightBeginJob struct {
}

// 挑战结束，发奖定时器
type WWWWeekFightEndJob struct {
}

// 数据清空定时器
type WWWWeekCleanJob struct {
}

// ------------------- 挑战前初始化定时器 -------------------

func (w *WWAWeekFightBeginJob) Run() {
	defer catchException()
	FightBeginImpl()
}

// 开始
func FightBeginImpl() {
	models.UserWWAWeekSwitch2Playoff()
}

// ------------------- 挑战结束定时器 -------------------

// 挑战结束，重置定时器

// 定时器
func (w *WWWWeekFightEndJob) Run() {
	defer catchException()
	models.SendUserBetResultMail()
}

// ------------------- 数据清空定时器 -------------------

func (w *WWWWeekCleanJob) Run() {
	defer catchException()
	models.WwaTypeForeach(models.ResetScore)
}
