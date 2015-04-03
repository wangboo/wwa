package mjob

import (
	// "github.com/revel/revel"
	"encoding/json"
	"fmt"
	"github.com/wangboo/wwa/app/models"
	"io/ioutil"
	"log"
	"net/http"
)

// 获取每日竞技人员
type RankDataJob struct {
}

func (r *RankDataJob) Run() {
	fmt.Println("Run RankData Job")
	models.DB.Exec("delete from rank")
	models.Redis.Do("DEL", "wwa_0")
	models.Redis.Do("DEL", "wwa_1")
	models.Redis.Do("DEL", "wwa_2")
	for _, s := range models.GameServerList {
		SaveDataByServerAndType(&s, 0)
		SaveDataByServerAndType(&s, 1)
		SaveDataByServerAndType(&s, 2)
	}
}

func SaveDataByServerAndType(s *models.GameServerConfig, t int) {
	resp, err := http.Get(s.UserRankUrl(t))
	if err != nil {
		log.Panicf("获取%s访问失败！！\n", s.UserRankUrl(t))
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	listOfRank := []models.Rank{}
	err = json.Unmarshal(data, &listOfRank)
	if err != nil {
		log.Panicln("SaveDataByServerAndType失败")
	}
	for index, r := range listOfRank {
		rank := &r
		rank.Score = 0
		rank.ZoneId = s.ZoneId
		rank.Type = t
		models.DB.Create(rank)
		models.Redis.Do("ZADD", rank.ToRedisRankName(), 0, rank.ToRedisRankValue())
		fmt.Printf("%d create rank %v \n", index, r)
	}
}
