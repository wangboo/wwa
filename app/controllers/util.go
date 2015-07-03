package controllers

import (
	"fmt"
	"github.com/revel/revel"
)

type AppsideController struct {
	revel.Controller
}

// 成功
func Succ(params ...interface{}) map[string]interface{} {
	lenOfParams := len(params)
	if lenOfParams%2 == 1 {
		panic("Succ入参必须为偶数")
	}
	rst := map[string]interface{}{}
	for i := 0; i < lenOfParams; i += 2 {
		key := params[i].(string)
		rst[key] = params[i+1]
	}
	fullfillAppsideMap(rst)
	return rst
}

func Fail(code, msg string) map[string]interface{} {
	rst := make(map[string]interface{}, 2)
	rst["_r"] = code
	rst["_m"] = msg
	return rst
}

func FailWithError(err error) map[string]interface{} {
	rst := make(map[string]interface{}, 2)
	rst["_r"] = "1"
	rst["_m"] = err.Error()
	return rst
}

func FailWithMsg(msg string, params ...interface{}) map[string]interface{} {
	rst := make(map[string]interface{}, 2)
	rst["_r"] = "1"
	rst["_m"] = fmt.Sprintf(msg, params...)
	return rst
}

func fullfillAppsideMap(rst map[string]interface{}) {
	if _, ok := rst["_m"]; !ok {
		rst["_m"] = ""
	}
	if _, ok := rst["_r"]; !ok {
		rst["_r"] = "0"
	}
}
