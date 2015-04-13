package mjob

import (
	"bufio"
	"encoding/base64"
	// "encoding/json"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/revel/revel"
	"github.com/wangboo/wwa/app/models"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type DayEndRewardJob struct {
}

// 奖励
type BaseReward struct {
	Type  int
	Begin int
	End   int
	Db    int
	Gold  int
}

func (b *BaseReward) String() string {
	return fmt.Sprintf("type:%d,(%d-%d) -- %d刀币，%d金币\n", b.Type, b.Begin, b.End, b.Db, b.Gold)
}

// 玩家奖励
type UserReward struct {
	UserId int `json:"u"`
	ZoneId int `json:"z"`
	Rank   int `json:"r"`
	Db     int `json:"d"`
	Gold   int `json:"g"`
}

func (u *UserReward) String() string {
	return fmt.Sprintf("%d,%d,%d,%d,%d", u.UserId, u.ZoneId, u.Rank, u.Db, u.Gold)
}

var (
	BaseRewardList []*BaseReward
)

// 日终发奖
func (j *DayEndRewardJob) Run() {
	loadConfig()
	path, _ := revel.Config.String("wwa.dayEndJobFile")
	filename := time.Now().Format("2006-01-02.txt")
	file, err := os.Create(fmt.Sprintf("%s/%s", path, filename))
	defer file.Close()
	if err != nil {
		log.Fatalf("创建文件出错：%s\n", err.Error())
	}
	dayEndRewardByType(0, file)
	dayEndRewardByType(1, file)
	dayEndRewardByType(2, file)
}

func loadConfig() {
	root, _ := revel.Config.String("root")
	filePath := fmt.Sprintf("%s/conf/base_wwa_reward.txt", root)
	file, err := os.Open(filePath)
	reader := bufio.NewReader(file)
	_, err = reader.ReadString('\n')
	if err != nil {
		log.Fatalf("加载 %s 出错，没有读取到第一行内容	\n", filePath)
		return
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		data := strings.Split(line, "\t")
		t, _ := strconv.Atoi(data[0])
		b, _ := strconv.Atoi(data[1])
		e, _ := strconv.Atoi(data[2])
		d, _ := strconv.Atoi(data[3])
		g, _ := strconv.Atoi(data[4])
		reward := &BaseReward{
			Type:  t,
			Begin: b,
			End:   e,
			Db:    d,
			Gold:  g,
		}
		BaseRewardList = append(BaseRewardList, reward)
		log.Printf("base :\n %v \n", BaseRewardList)
	}
}

func dayEndRewardByType(Type int, file *os.File) {
	cli := models.RedisPool.Get()
	defer cli.Close()
	TypeName := fmt.Sprintf("wwa_%d", Type)
	maxScore := fmt.Sprintf("(%d", models.RANK_SCORE_SUB)
	rankUsers, _ := redis.Strings(cli.Do("ZRANGEBYSCORE", TypeName, "-inf", maxScore))
	// log.Printf("rankUsers = %v \n", rankUsers)
	size := len(rankUsers)
	if size == 0 {
		return
	}
	list := make([]*UserReward, size)
	for i := 0; i < size; i++ {
		rank := i + 1
		data := strings.Split(rankUsers[i], ",")
		base, err := getRewardByRank(Type, rank)
		if err != nil {
			log.Fatal("error : %s \n", err.Error())
			continue
		}
		userId, _ := strconv.Atoi(data[1])
		zoneId, _ := strconv.Atoi(data[0])
		user := &UserReward{
			UserId: userId,
			ZoneId: zoneId,
			Rank:   rank,
			Db:     base.Db,
			Gold:   base.Gold,
		}
		list[i] = user
	}
	// 发奖
	for _, gs := range models.GameServerList {
		users := findUserByGameServer(list, &gs)
		str := strings.Join(users, "-")
		// log.Printf("发奖给%s:\n%s\n", gs.Name, str)
		encode := base64.StdEncoding.EncodeToString([]byte(str))
		// 下发日终奖励
		go func() {
			// log.Printf("go %s\n", gs.DayEndWwaRewardUrl(encode, Type))
			ok, err := models.GetGameServer(gs.DayEndWwaRewardUrl(encode, Type))
			if err != nil {
				log.Printf("访问游戏服务器出错\n")
			}
			log.Printf("DayEndRewardJob resp %s \n", ok)
		}()
		file.WriteString(fmt.Sprintf("### %d组 发奖给%d服-%s \n", Type, gs.ZoneId, gs.Name))
		file.WriteString(strings.Join(users, "\n"))
		file.WriteString("\n\n")
	}
}

func getRewardByRank(t, rank int) (*BaseReward, error) {
	for _, r := range BaseRewardList {
		if r.Type == t && r.Begin <= rank && r.End >= rank {
			return r, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("找不到基本排名奖励 type=%d, rank=%d\n", t, rank))
}

// 从全部玩家中找到一个服务器的玩家
func findUserByGameServer(list []*UserReward, gs *models.GameServerConfig) []string {
	all := []string{}
	for _, u := range list {
		if u.ZoneId == gs.ZoneId {
			all = append(all, u.String())
		}
	}
	return all
}
