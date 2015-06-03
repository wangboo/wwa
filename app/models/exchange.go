package models

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/revel/revel"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Exchange struct {
	Id          int    `json:"id"`
	Name        string `json:"n"`
	Dec         string `json:"d"`
	Reward      string `json:"r"`
	Money       int    `json:"m"`
	DayTimesLim int    `json:"l"`
	Rates       int    `json:"-"`
}

// 奖励预览
type ExchangeView struct {
	Quanlity int `json:"q"`
	Img      int `json:"img"`
	Amt      int `json:"amt"`
}

var (
	// 基础兑换表
	BaseExchangeList = []*Exchange{}
	// 今日兑换内容
	TodayExchangeList = []*Exchange{}
	// 每日兑换在redis中key
	TodayExchangeKey = "TodayExchangeKey"
)

func ReloadExchangeConfig() error {
	BaseExchangeList = []*Exchange{}
	root, _ := revel.Config.String("root")
	file, err := os.Open(fmt.Sprintf("%s/conf/base_wwa_exchange.txt", root))
	defer file.Close()
	if err != nil {
		return err
	}
	reader := bufio.NewReader(file)
	_, err = reader.ReadString('\n')
	if err != nil {
		return errors.New(fmt.Sprintf("读取文件首行出错:%s", err.Error()))
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		line = strings.TrimSpace(line)
		ss := strings.Split(line, "\t")
		Id, err := strconv.Atoi(ss[0])
		if err != nil {
			return errors.New(fmt.Sprintf("读取 Id 出错:%s", err.Error()))
		}
		money, err := strconv.Atoi(ss[4])
		if err != nil {
			return errors.New(fmt.Sprintf("读取 Money 出错:%s", err.Error()))
		}
		dayTimesLim, err := strconv.Atoi(ss[5])
		if err != nil {
			return errors.New(fmt.Sprintf("读取 DayTimesLim 出错:%s", err.Error()))
		}
		rates, err := strconv.Atoi(ss[6])
		if err != nil {
			return errors.New(fmt.Sprintf("读取 Rates 出错:%s", err.Error()))
		}
		ex := &Exchange{
			Id:          Id,
			Name:        ss[1],
			Dec:         ss[2],
			Reward:      ss[3],
			Money:       money,
			DayTimesLim: dayTimesLim,
			Rates:       rates,
		}
		BaseExchangeList = append(BaseExchangeList, ex)
	}
	return nil
}

func ResetDailyExchange() error {
	TodayExchangeList = []*Exchange{}
	if len(BaseExchangeList) == 0 {
		return errors.New("基本兑换表 base_wwa_exchange 还没有加载或者没有内容")
	}
	sumRate := 0
	for _, v := range BaseExchangeList {
		sumRate += v.Rates
	}
	amt := revel.Config.IntDefault("exchange.amount", 6)
	max := len(BaseExchangeList)
	if amt > max {
		amt = max
	}
	rand.Seed(time.Now().Unix())
	for {
		randNum := rand.Intn(sumRate)
		ex := getExchangeByRates(randNum)
		if exNotInTodayExchangeList(ex) {
			TodayExchangeList = append(TodayExchangeList, ex)
			if len(TodayExchangeList) >= amt {
				break
			}
		}
	}
	// 将结果存放在redis中备份
	cli := RedisPool.Get()
	defer cli.Close()
	jsonData, _ := json.Marshal(TodayExchangeList)
	log.Printf("save TodayExchangeList : %s\n", jsonData)
	cli.Do("set", TodayExchangeKey, jsonData)
	return nil
}

// ex是否已经在 TodayExchangeList 中
func exNotInTodayExchangeList(ex *Exchange) bool {
	for _, v := range TodayExchangeList {
		if v.Id == ex.Id {
			return false
		}
	}
	return true
}

func getExchangeByRates(rand int) *Exchange {
	for _, v := range BaseExchangeList {
		if rand < v.Rates {
			return v
		}
		rand -= v.Rates
	}
	return nil
}
