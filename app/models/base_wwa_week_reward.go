package models

import (
	"bufio"
	// "bytes"
	"fmt"
	"github.com/revel/revel"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

//巅峰竞技奖励
type BaseWWAWeekReward struct {
	Type   int
	Begin  int
	End    int
	Reward string
}

var BaseWWAWeekRewardList []*BaseWWAWeekReward

func LoadBaseWWAWeekReward() {
	BaseWWAWeekRewardList = []*BaseWWAWeekReward{}
	rootPath, _ := revel.Config.String("root")
	filePath := path.Join(rootPath, "conf", "base_wwa_week_reward.txt")
	file, err := os.Open(filePath)
	if err != nil {
		panic(fmt.Sprintf("读取配置文件%s报错：%s", filePath, err.Error()))
	}
	buffer := bufio.NewReader(file)
	_, err = buffer.ReadString('\n')
	if err != nil {
		panic(fmt.Sprintf("读取配置文件首行出错：%s", err.Error()))
	}
	for {
		line, err := buffer.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(fmt.Sprintf("读取配置文件出错：%s", err.Error()))
			}
		}
		// 读取配置文件
		ss := strings.Split(line, "\t")
		if len(ss) < 4 {
			panic("配置文件base_wwa_week_reward不正确，配置项不足4项")
		}
		item := &BaseWWAWeekReward{}

		var intVal = 0
		intVal, err = strconv.Atoi(ss[0])
		if err != nil {
			revel.ERROR.Println("BaseWWAWeekReward读取type出错，line：", line)
		} else {
			item.Type = intVal
		}

		intVal, err = strconv.Atoi(ss[1])
		if err != nil {
			revel.ERROR.Println("BaseWWAWeekReward读取begin出错，line：", line)
		} else {
			item.Begin = intVal
		}

		intVal, err = strconv.Atoi(ss[2])
		if err != nil {
			revel.ERROR.Println("BaseWWAWeekReward读取End出错，line：", line)
		} else {
			item.End = intVal
		}

		item.Reward = ss[3]

		BaseWWAWeekRewardList = append(BaseWWAWeekRewardList, item)
	}
}

// 巅峰之夜根据排名获取奖励
func BaseWWAWeekRewardGetRewardByTypeAndRank(typeOfWwa, rank int) string {
	for _, item := range BaseWWAWeekRewardList {
		if item.Type == typeOfWwa && item.Begin <= rank && item.End >= rank {
			return item.Reward
		}
	}
	return ""
}
