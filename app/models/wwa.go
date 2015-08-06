package models

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"strings"
)

// zone_user 存放value的对象
type WWA []string

// return fmt.Sprintf("%d,%d,%d,%s,%d,%d,%d,%d,%s,%d,%d,%d",
// r.UserId, r.Score, r.Level, r.Name,
// r.Pow, r.Hero, r.Q, r.ZoneId,
// r.ZoneName, r.Type, r.Img, r.Frame)

// 以字符串创建WWA对象
func CreateWWAWithString(desc string) WWA {
	list := strings.Split(desc, ",")
	return CreateWWAWithArray(list)
}

// 以数组创建WWA对象
func CreateWWAWithArray(list []string) WWA {
	if len(list) == 12 {
		return WWA(list)
	}
	panic(fmt.Sprintf("desc = %v not length of 12", list))
}

func FindWWAInRedis(zoneId, userId int) (wwa WWA, err error) {
	cli := RedisPool.Get()
	defer cli.Close()
	detail, err := redis.String(cli.Do("HGET", "zone_user", ToSimpleKey(zoneId, userId)))
	if err != nil {
		return
	}
	wwa = CreateWWAWithString(detail)
	return
}

// 跨服竞技玩家id
func (w WWA) UserId() (val int) {
	val, _ = strconv.Atoi(w[0])
	return
}

// 跨服竞技积分
func (w WWA) Score() (val int) {
	val, _ = strconv.Atoi(w[1])
	return
}

// 设置跨服竞技积分
func (w WWA) SetScore(score int) {
	w[1] = strconv.Itoa(score)
}

// 跨服竞技等级
func (w WWA) Level() (val int) {
	val, _ = strconv.Atoi(w[2])
	return
}

// 跨服竞技名字
func (w WWA) Name() string {
	return w[3]
}

// 跨服竞技战斗力
func (w WWA) Pow() (val int) {
	val, _ = strconv.Atoi(w[4])
	return
}

// 跨服竞技英雄头像
func (w WWA) Hero() (val int) {
	val, _ = strconv.Atoi(w[5])
	return
}

// 跨服竞技英雄品质
func (w WWA) Q() (val int) {
	val, _ = strconv.Atoi(w[6])
	return
}

// 区id
func (w WWA) ZoneId() (val int) {
	val, _ = strconv.Atoi(w[7])
	return
}

// 区id
func (w WWA) ZoneName() string {
	return w[8]
}

// 区竞技场类别0，1，2
func (w WWA) Type() (val int) {
	val, _ = strconv.Atoi(w[9])
	return
}

// 区竞技场头像
func (w WWA) Img() (val int) {
	val, _ = strconv.Atoi(w[10])
	return
}

// 区竞技场头像框
func (w WWA) Frame() (val int) {
	val, _ = strconv.Atoi(w[11])
	return
}

func (w WWA) String() string {
	return strings.Join(w, ",")
}

func (w WWA) SimpleKey() string {
	return ToSimpleKey(w.ZoneId(), w.UserId())
}

func (w WWA) UpdateToRedis() {
	cli := RedisPool.Get()
	defer cli.Close()
	//	更新缓存数据
	cli.Do("HSET", "zone_user", w.SimpleKey(), w.String())
}
