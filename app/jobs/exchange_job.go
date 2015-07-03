package mjob

import (
	"github.com/wangboo/wwa/app/models"
	"log"
)

type ExchangeJob struct {
}

func (e *ExchangeJob) Run() {
	defer catchException()
	e.RunImpl()
}

func (e *ExchangeJob) RunImpl() {
	if len(models.BaseExchangeList) == 0 {
		err := models.ReloadExchangeConfig()
		if err != nil {
			log.Printf("ExchangeJob 载入基本表出错！%s\n", err.Error())
			return
		}
	}
	err := models.ResetDailyExchange()
	if err != nil {
		log.Printf("ExchangeJob 重置每日兑换报错:%s\n", err.Error())
	}
}
