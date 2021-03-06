package controllers

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	// "github.com/revel/modules/jobs/app/jobs"
	"encoding/base64"
	"github.com/revel/revel"
	"github.com/wangboo/wwa/app/jobs"
	"github.com/wangboo/wwa/app/models"
	"math/rand"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	GetNameRegex = regexp.MustCompile(`^(\d+,){3}(.*?),.*`)
)

type ArenaCtrl struct {
	*revel.Controller
}

// DetailKey排序
type SortDetailByLevel []string

func (s SortDetailByLevel) Len() int {
	return len(s)
}

func (s SortDetailByLevel) Swap(a, b int) {
	s[a], s[b] = s[b], s[a]
}

func (s SortDetailByLevel) Less(a, b int) bool {
	intA, _ := strconv.Atoi(strings.Split(s[a], ",")[2])
	intB, _ := strconv.Atoi(strings.Split(s[b], ",")[2])
	return intA >= intB
}

func (c ArenaCtrl) GroupData(a, u int) revel.Result {
	return c.RenderText(GetGroupData(a, u))
}

func GetGroupData(a, u int) string {
	cli := models.RedisPool.Get()
	defer cli.Close()
	gi, err := cli.Do("GET", models.GroupDataKey(u, a))
	if err != nil {
		return fmt.Sprintf("redis err %v", err)
	}
	if gi != nil {
		return fmt.Sprintf("%s", gi)
	}
	str := models.GetGroupDataFromGameServer(a, u)
	if len(str) == 0 {
		return ""
	}
	sec := revel.Config.IntDefault("group_data_cache_time", 60)
	cli.Do("SETEX", models.GroupDataKey(u, a), sec, str)
	return str
}

func (c ArenaCtrl) TopFifty(id int) revel.Result {
	cli := models.RedisPool.Get()
	defer cli.Close()
	allSimplyKeys, _ := redis.Strings(cli.Do("ZRANGE", fmt.Sprintf("wwa_%d", id), 0, 49))
	if len(allSimplyKeys) == 0 {
		return c.RenderText("")
	}
	args := redis.Args{}
	args = args.Add("zone_user")
	for _, k := range allSimplyKeys {
		args = args.Add(k)
	}
	// log.Printf("args = %v \n", args)
	all, err := redis.Strings(cli.Do("HMGET", args...))
	if err != nil {
		return c.RenderText("redis err %v", err.Error())
	}
	return c.RenderText(strings.Join(all, "-"))
}

// 获得积分 或者 消费积分
func (c ArenaCtrl) FightResult(s, u, a, us, uu, ua int, win bool) revel.Result {
	c.Validation.Required(s).Message("您必须提供 s:获得积分")
	c.Validation.Required(u).Message("您必须提供 u:用户编号")
	c.Validation.Required(a).Message("您必须提供 a:服务器编号")
	if c.Validation.HasErrors() {
		return RenderValidationFail(c.Controller)
	}
	cli := models.RedisPool.Get()
	defer cli.Close()
	record, err := incrScore(cli, a, u, s)
	if err != nil {
		return c.RenderText("redis err %s", err.Error())
	}
	// log.Printf("win = %v\n", win)
	if win && uu > 0 {
		// 对方扣分
		if _, err := incrScore(cli, ua, uu, -us); err != nil {
			return c.RenderText("redis err %s", err.Error())
		}
		server := models.FindGameServer(ua)
		// rst := GetNameRegex.FindStringSubmatch(myDetail)
		content := fmt.Sprintf("你在跨服竞技中遭遇%s的突袭，将军被迫撤退，损失了%d点竞技积分。", record.Name(), us)
		content = base64.StdEncoding.EncodeToString([]byte(content))
		content = url.QueryEscape(content)
		go models.GetGameServer(server.MailUrl(uu, content))
	}
	return c.RenderText("ok")
}

// 玩家增加/扣除积分
func incrScore(cli redis.Conn, a, u, s int) (record models.WWA, err error) {
	record, err = models.FindWWAInRedis(a, u)
	if err != nil {
		return
	}
	newScore := record.Score() + s
	if newScore < 0 {
		newScore = 0
	}
	revel.INFO.Println("newScore = ", newScore)
	record.SetScore(newScore)
	// detail, err := redis.String(cli.Do("HGET", "zone_user", models.ToSimpleKey(a, u)))
	// if err != nil {
	// 	return "", err
	// }
	// revel.INFO.Printf("incrScore before detail : %s\n", detail)
	// rst := strings.Split(detail, ",")
	// score, _ := strconv.Atoi(rst[1])
	// rst[1] = strconv.Itoa(newScore)
	// newStr := strings.Join(rst, ",")
	// revel.INFO.Println("incrScore after : ", newStr)
	// simpleKey := models.ToSimpleKey(a, u)
	//	更新缓存数据
	// cli.Do("HSET", "zone_user", simpleKey, newStr)
	// revel.INFO.Println("record.Score = ", record.Score())
	record.UpdateToRedis()
	simpleKey := record.SimpleKey()
	wwa := fmt.Sprintf("wwa_%d", record.Type())
	// rankScore, _ := redis.Int(cli.Do("ZSCORE", wwa, simpleKey))
	// revel.INFO.Printf("wwa = %s, simpleKey = %s, rankScore = %d \n", wwa, simpleKey, rankScore)
	// rankScore = rankScore - s
	// if rankScore > models.RANK_SCORE_SUB {
	// 	rankScore = models.RANK_SCORE_SUB
	// }
	rankScore := models.RANK_SCORE_SUB - newScore
	cli.Do("ZADD", wwa, strconv.Itoa(rankScore), simpleKey)
	// 巅峰之夜积分变化
	// revel.INFO.Println("巅峰之夜积分变化")
	models.UserWWAWeekScoreChange(a, u, s, record.Type())
	return
}

// 队伍信息
func (c ArenaCtrl) GroupInfo(u, a int) revel.Result {
	return c.RenderText(GetGroupInfo(u, a))
}

func GetGroupInfo(u, a int) string {
	cli := models.RedisPool.Get()
	defer cli.Close()
	gi, err := cli.Do("GET", models.GroupInfoKey(u, a))
	if err != nil {
		fmt.Sprintf("redis err %v", err)
	}
	if gi != nil {
		return fmt.Sprintf("%s", gi)
	}
	str := models.GetGroupInfoFromGameServer(a, u)
	if len(str) == 0 {
		return ""
	}
	sec := revel.Config.IntDefault("group_info_cache_time", 60)
	cli.Do("SETEX", models.GroupInfoKey(u, a), sec, str)
	return str
}

func (c *ArenaCtrl) MuInfo(u, a int) revel.Result {
	cli := models.RedisPool.Get()
	defer cli.Close()
	gi, err := cli.Do("GET", models.MuInfoKey(u, a))
	if err != nil {
		c.RenderText("redis err %v", err)
	}
	if gi != nil {
		return c.RenderText("%s", gi)
	}
	str := models.GetMuInfoFromGameServer(a, u)
	if len(str) == 0 {
		return c.RenderText("")
	}
	sec := revel.Config.IntDefault("group_info_cache_time", 60)
	cli.Do("SETEX", models.MuInfoKey(u, a), sec, str)
	return c.RenderText(str)
}

// 基本信息
func (c ArenaCtrl) BaseInfo(u, a int) revel.Result {
	cli := models.RedisPool.Get()
	defer cli.Close()
	detail, err := cli.Do("HGET", "zone_user", models.ToSimpleKey(a, u))
	if err != nil {
		return c.RenderText("redis error %v", err.Error())
	}
	return c.RenderText("%s", detail)
}

func (c ArenaCtrl) RankInfo(u, a int) revel.Result {
	c.Validation.Required(u).Message("您必须提供 u:用户编号")
	c.Validation.Required(a).Message("您必须提供 a:服务器编号")
	if c.Validation.HasErrors() {
		return RenderValidationFail(c.Controller)
	}
	cli := models.RedisPool.Get()
	defer cli.Close()
	simpleKey := models.ToSimpleKey(a, u)
	detail, err := redis.String(cli.Do("HGET", "zone_user", simpleKey))
	if err != nil {
		return c.RenderText("redis error %s", err.Error())
	}
	rst := strings.Split(detail, ",")
	rank, _ := cli.Do("ZRANK", fmt.Sprintf("wwa_%s", rst[9]), simpleKey)
	// type,rank,积分
	return c.RenderText("%s,%d,%s", rst[9], rank, rst[1])
}

func (c ArenaCtrl) ResetRank(pwd string) revel.Result {
	if pwd != revel.Config.StringDefault("password", "w231520") {
		c.RenderText("fail password error")
	}
	(&mjob.RankDataJob{}).Run()
	return c.RenderText("success")
}

func (c ArenaCtrl) DayEndRewardJob(pwd string) revel.Result {
	revel.INFO.Println("DayEndRewardJob trigger from web")
	if pwd != revel.Config.StringDefault("password", "w231520") {
		c.RenderText("fail password error")
	}
	(&mjob.DayEndRewardJob{}).Run()
	return c.RenderText("success")
}

func (c ArenaCtrl) SendDayEndRewardByType(t int) revel.Result {
	job := &mjob.DayEndRewardJob{}
	job.SendRewardByType(t)
	return c.RenderText("ok")
}

// 随机3名挑战对象
func (c ArenaCtrl) RandFightUsers(u, a int) revel.Result {
	revel.INFO.Println("RandFightUsers")
	cli := models.RedisPool.Get()
	defer cli.Close()
	revel.INFO.Println("GET Redis")
	simpleKey := models.ToSimpleKey(a, u)
	detail, err := redis.String(cli.Do("HGET", "zone_user", simpleKey))
	if err != nil {
		return c.RenderText("redis error", err.Error())
	}
	revel.INFO.Printf("myself : %v \n", detail)
	rst := strings.Split(detail, ",")
	wwa := fmt.Sprintf("wwa_%s", rst[9])
	size, _ := redis.Int(cli.Do("ZCard", wwa))
	if size < 4 {
		revel.INFO.Println("player not greater than 4")
		return c.RenderText("redis error")
	}
	rank, _ := redis.Int(cli.Do("ZRANK", wwa, simpleKey))
	ranks := make([]int, 3)
	if size < 40 {
		randNum := rand.Perm(size)
		for i, j := 0, 0; i < 3; {
			revel.INFO.Printf("i=%d,randNum[i]=%d,rank=%d\n", i, randNum[i], rank)
			if randNum[j] != rank {
				ranks[i] = randNum[j]
				i += 1
			}
			j += 1
		}
	} else {
		// 最高经验
		if rank < 20 {
			ranks[0] = getRandExcept(0, 10, rank)
		} else {
			ranks[0] = getRandExcept(rank-11, rank-1, rank)
		}
		// 中等
		if ranks[0]+11 > size {
			ranks[1] = getRandExcept(size-10, size-1, 0)
		} else {
			ranks[1] = getRandExcept(ranks[0]+1, ranks[0]+10, rank)
		}
		// 低等
		if ranks[1]+20 > size {
			ranks[2] = getRandExcept(size-15, size-5, 0)
		} else {
			ranks[2] = getRandExcept(ranks[1]+1, ranks[1]+15, rank)
		}
	}
	fmt.Printf("ranks = %v, rank=%d, size=%d \n", ranks, rank, size)
	rst0, _ := redis.Strings(cli.Do("ZRANGE", wwa, ranks[0], ranks[0]))
	rst1, _ := redis.Strings(cli.Do("ZRANGE", wwa, ranks[1], ranks[1]))
	rst2, _ := redis.Strings(cli.Do("ZRANGE", wwa, ranks[2], ranks[2]))
	repl, _ := redis.Strings(cli.Do("HMGET", "zone_user", rst0[0], rst1[0], rst2[0]))
	// log.Printf("repl : %v \n", repl)
	sortVal := SortDetailByLevel(repl)
	sort.Sort(sortVal)
	return c.RenderText(strings.Join(sortVal, "-"))
}

func getRandExcept(start, end, except int) int {
	mRange, n := end-start, 0
	for {
		if n = start + rand.Intn(mRange); n != except {
			return n
		}
	}
}

func getThreeExcept(ranks []string, except string) []string {
	size := len(ranks)
	rst := []string{"", "", ""}
	index := 0
	num := ""
	for {
		num = ranks[rand.Intn(size)]
		if rst[0] == num || rst[1] == num || rst[2] == num || num == except {
			continue
		}
		rst[index] = num
		if index == 2 {
			break
		}
		index += 1
	}
	return rst
}

// 新升级的用户
func (c ArenaCtrl) NewComer(a, u, pow, hero, q, lev, img, frame int, name string) revel.Result {
	cli := models.RedisPool.Get()
	defer cli.Close()
	gs := models.FindGameServer(a)
	if gs == nil {
		revel.ERROR.Printf("找不到游戏服务器：arenId=%d, userId=%d, name=%s \n", a, u, name)
		return c.RenderText("ok")
	}
	rank := &models.Rank{
		UserId:   u,
		Score:    0,
		Level:    lev,
		Name:     name,
		Hero:     hero,
		Q:        q,
		Pow:      pow,
		ZoneId:   a,
		Type:     0,
		Img:      img,
		Frame:    frame,
		ZoneName: gs.Name,
	}
	revel.INFO.Printf("newComer name = %s, Frame=%d,Img=%d\n", rank.Name, rank.Frame, rank.Img)
	simpleKey := models.ToSimpleKey(a, u)
	cli.Do("HSET", "zone_user", simpleKey, rank.ToDetailKey())
	cli.Do("ZADD", "wwa_0", models.RANK_SCORE_SUB, simpleKey)
	return c.RenderText("ok")
}
