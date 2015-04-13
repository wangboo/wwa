package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	now := time.Now()
	filename := now.Format("2006-01-02.txt")
	file, err := os.Create(fmt.Sprintf("/Users/wangboo/Desktop/%s", filename))
	if err != nil {
		fmt.Printf("err %s\n", err.Error())
	}
	file.WriteString("#### fuck")
	file.WriteString("#### fuck")
	defer file.Close()
}
