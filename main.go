package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
)

func main() {
	con, err := redis.Dial("tcp", ":6379")
	if err != nil {
		panic(fmt.Sprintf("连接redis错误 %s \n", err.Error()))
	}
	users, _ := redis.Strings(con.Do("ZRANGE", "wwa_0", 0, 49))
	fmt.Println("len", len(users))
	// ranks := []Rank{}
	// redis.ScanSlice(all, &ranks)
	// fmt.Printf("name = %v \n", ranks)
}
