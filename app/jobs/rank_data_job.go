package mjob

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/revel/revel"
	"github.com/wangboo/wwa/app/models"
	"io/ioutil"
	"net/http"
)

// 获取每日竞技人员
type RankDataJob struct {
}

func (r *RankDataJob) Run() {
	cli := models.RedisPool.Get()
	defer cli.Close()
	// models.DB.Exec("delete from ranks")
	// wwa_ 为 Sorted-set，存放着玩家竞技排名信息 score - 玩家详细信息
	cli.Do("DEL", "wwa_0")
	cli.Do("DEL", "wwa_1")
	cli.Do("DEL", "wwa_2")
	// zone_user 为hash表，存放着k-v 内容为 (服务器编号-玩家编号)-(玩家详细信息)
	cli.Do("DEL", "zone_user")
	for _, s := range models.GameServerList {
		SaveDataByServerAndType(cli, &s, 0)
		SaveDataByServerAndType(cli, &s, 1)
		SaveDataByServerAndType(cli, &s, 2)
	}
}

func SaveDataByServerAndType(cli redis.Conn, s *models.GameServerConfig, t int) {
	url := s.UserRankUrl(t)
	resp, err := http.Get(url)
	revel.INFO.Printf("url : %s\n", url)
	if err != nil {
		revel.ERROR.Printf("获取%s访问失败！！\n", url)
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	revel.INFO.Printf("服务器应答：%s \n", string(data))
	listOfRank := []models.Rank{}
	err = json.Unmarshal(data, &listOfRank)
	if err != nil {
		revel.ERROR.Printf("SaveDataByServerAndType失败\n", data)
		return
	}
	for _, r := range listOfRank {
		rank := &r
		revel.INFO.Println("rank = %v ", rank)
		rank.Score = 0
		rank.ZoneId = s.ZoneId
		rank.ZoneName = s.Name
		rank.Type = t
		cli.Do("ZADD", rank.ToRedisRankName(), models.RANK_SCORE_SUB, rank.ToSimpleKey())
		cli.Do("HSET", "zone_user", rank.ToSimpleKey(), rank.ToDetailKey())
	}
}
