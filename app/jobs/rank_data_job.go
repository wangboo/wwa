package mjob

import (
	"encoding/json"
	// "fmt"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/revel/revel"
	"github.com/wangboo/wwa/app/models"
	"io/ioutil"
	"net/http"
	"time"
)

// 获取每日竞技人员
type RankDataJob struct {
}

func (r *RankDataJob) Run() {
	defer catchException()
	r.RunImpl()
}

func (r *RankDataJob) RunImpl() {
	cli := models.RedisPool.Get()
	defer cli.Close()
	// models.DB.Exec("delete from ranks")
	// wwa_ 为 Sorted-set，存放着玩家竞技排名信息 score - 玩家详细信息
	// cli.Do("DEL", "wwa_0")
	// cli.Do("DEL", "wwa_1")
	// cli.Do("DEL", "wwa_2")
	// cli.Do("DEL", "wwa_3")
	models.WwaTypeForeach(func(typeOfWwa int) {
		cli.Do("DEL", fmt.Sprintf("wwa_%d", typeOfWwa))
	})
	// zone_user 为hash表，存放着k-v 内容为 (服务器编号,玩家编号)-(玩家详细信息)
	cli.Do("DEL", "zone_user")
	for _, s := range models.GameServerList {
		models.WwaTypeForeach(func(typeOfWwa int) {
			SaveDataByServerAndType(cli, &s, typeOfWwa)
		})
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
	defer resp.Body.Close()
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
		rank.Score = 0
		rank.ZoneId = s.ZoneId
		rank.ZoneName = s.Name
		rank.Type = t
		week, err := models.FindWWAWeekByZoneIdAndUserId(s.ZoneId, rank.UserId)
		if err == nil && time.Now().Weekday() < 5 && time.Now().Weekday() > 0 {
			week.Type = t
			week.Pow = rank.Pow
			week.Save()
		}
		// revel.INFO.Println("rank = ", rank)
		cli.Do("ZADD", rank.ToRedisRankName(), models.RANK_SCORE_SUB, rank.ToSimpleKey())
		cli.Do("HSET", "zone_user", rank.ToSimpleKey(), rank.ToDetailKey())
	}
}
