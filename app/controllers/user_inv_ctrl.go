package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/revel/revel"
	"github.com/wangboo/wwa/app/models"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
)

type UserInvCtrl struct {
	*revel.Controller
	MSession *mgo.Session
}

func (c *UserInvCtrl) Code(username string, zoneId int) revel.Result {
	col := models.Col(c.MSession, models.COL_USER_INV)
	user, err := models.FindUserInv(col, username, zoneId)
	if err != nil {
		if err == mgo.ErrNotFound {
			user, err = models.NewUserInv(col, username, zoneId)
		} else {
			return c.RenderJson(FailWithError(err))
		}
	}
	return c.RenderJson(bson.M{"code": user.Code})
}

// 可领取的任务数量
func (c *UserInvCtrl) ShowSize(username string, zoneId int) revel.Result {
	col := models.Col(c.MSession, models.COL_USER_INV)
	user, err := models.FindUserInv(col, username, zoneId)
	if err != nil {
		return c.RenderJson(FailWithError(err))
	}
	count := 0
	// 每日任务
	for _, baseDaily := range models.BaseDailyInvList {
		ud, ok := findUserDailyTaskByTaskId(baseDaily.Id, user.DailyTask)
		if !ok {
			revel.ERROR.Printf("not find daily task! user=%s, zoneId=%d, taskId=%d \n", username, zoneId, baseDaily.Id)
		}
		if ud.Got {
			continue
		}
		if ok && ud.Complete {
			count += 1
		}
	}
	// 主线任务
	focusUsers := []models.UserInv{}
	col.Find(bson.M{"focus": user.Id}).All(&focusUsers)
	userTaskInfos := models.FindUserTaskInfos(focusUsers)
	needUpdate := false
	for index, baseMain := range models.BaseMainInvList {
		um, ok := findUserDailyTaskByTaskId(baseMain.Id, user.MainTask)
		if !ok || um.Got {
			continue
		}
		// 检查任务是否完成
		if um.Complete {
			count += 1
		} else {
			if userInvMainTaskIsComplete(&baseMain, userTaskInfos) {
				// 需要将更新状态
				user.MainTask[index].Complete = true
				needUpdate = true
				count += 1
			}
		}
	}
	if needUpdate {
		revel.INFO.Println("update user.MainTask = ", user.MainTask)
		col.UpdateId(user.Id, bson.M{"$set": bson.M{"main_task": user.MainTask}})
	}
	return c.RenderJson(bson.M{"size": count})
}

func (c *UserInvCtrl) Show(username string, zoneId int) revel.Result {
	col := models.Col(c.MSession, models.COL_USER_INV)
	user, err := models.FindUserInv(col, username, zoneId)
	if err != nil {
		return c.RenderJson(FailWithError(err))
	}
	rst := bson.M{}
	tasks := []interface{}{}
	// 每日任务
	for _, baseDaily := range models.BaseDailyInvList {
		ud, ok := findUserDailyTaskByTaskId(baseDaily.Id, user.DailyTask)
		if !ok {
			revel.ERROR.Printf("not find daily task! user=%s, zoneId=%d, taskId=%d \n", username, zoneId, baseDaily.Id)
		}
		daily := bson.M{}
		daily["reward"] = baseDaily.Reward
		daily["complete"] = ok && ud.Complete
		daily["got"] = ud.Got
		daily["desc"] = baseDaily.Desc
		daily["taskId"] = baseDaily.Id
		tasks = append(tasks, daily)
	}
	// 主线任务
	focusUsers := []models.UserInv{}
	col.Find(bson.M{"focus": user.Id}).All(&focusUsers)
	userTaskInfos := models.FindUserTaskInfos(focusUsers)
	needUpdate := false
	for index, baseMain := range models.BaseMainInvList {
		um, ok := findUserDailyTaskByTaskId(baseMain.Id, user.MainTask)
		if !ok {
			continue
		}
		main := bson.M{}
		main["got"] = um.Got
		main["reward"] = baseMain.Reward
		// 检查任务是否完成
		if um.Complete {
			main["complete"] = true
		} else {
			if userInvMainTaskIsComplete(&baseMain, userTaskInfos) {
				// 需要将更新状态
				main["complete"] = true
				user.MainTask[index].Complete = true
				needUpdate = true
			} else {
				main["complete"] = false
			}
		}
		main["desc"] = baseMain.Desc
		main["taskId"] = baseMain.Id
		tasks = append(tasks, main)
	}
	if needUpdate {
		revel.INFO.Println("update user.MainTask = ", user.MainTask)
		col.UpdateId(user.Id, bson.M{"$set": bson.M{"main_task": user.MainTask}})
	}
	rst["tasks"] = tasks
	return c.RenderJson(rst)
}

// // 玩家升级 从from等级到to等级
func (c *UserInvCtrl) LevelUp(username string, zoneId int, from, to int) revel.Result {
	col := models.Col(c.MSession, models.COL_USER_INV)
	user, err := models.FindUserInv(col, username, zoneId)
	if err != nil {
		return c.RenderJson(FailWithError(err))
	}
	focus, err := models.FindUserInvByFocusId(col, user.Id)
	if err != nil {
		revel.INFO.Println("FindUserByFocusId err")
		return c.RenderJson(FailWithError(err))
	}
	return c.RenderJson(focus)
}

// 使用激活码
func (c *UserInvCtrl) UseCode(username string, zoneId int, code string) revel.Result {
	col := models.Col(c.MSession, models.COL_USER_INV)
	user, err := models.FindUserInv(col, username, zoneId)
	if err != nil {
		revel.INFO.Println("err = ", err)
		return c.RenderJson(FailWithMsg("找不到玩家: zoneId=%d, username=%s", zoneId, username))
	}
	revel.INFO.Println("user = ", user.Focus)
	if user.Focus.Valid() {
		return c.RenderJson(FailWithMsg("您已经使用过推广码了"))
	}
	focus, err := models.FindUserInvByCode(col, code)
	if err != nil {
		return c.RenderJson(FailWithMsg("找不到推广码:%s", code))
	}
	if user.Id == focus.Id {
		return c.RenderJson(FailWithMsg("您不能使用自己的推广码"))
	}
	err = col.UpdateId(user.Id, bson.M{"$set": bson.M{"focus": focus.Id}})
	// focus 方每日任务检查
	needUpdate := false
	if taskStatusCompletionSize(user.DailyTask) < len(models.BaseDailyInvList) {
		revel.INFO.Println("检查DailyTask")
		// 还没有完成每日任务
		for _, bd := range models.BaseDailyInvList {
			ud, _ := findUserDailyTaskByTaskId(bd.Id, focus.DailyTask)
			if ud.Got {
				// 任务已经完成
				continue
			}
			if models.ArrayContainObjectId(ud.CompletionIds, user.Id) {
				revel.INFO.Printf("user %s already in ud.CompletionIds %v \n", user.Id.Hex(), ud.CompletionIds)
				// 已经在列表中了
				continue
			}
			ud.CompletionIds = append(ud.CompletionIds, user.Id)
			revel.INFO.Printf("append %s to ud.CompletionIds %v \n", user.Id.Hex(), ud.CompletionIds)
			needUpdate = true
			if len(ud.CompletionIds) >= bd.Size {
				// 完成
				revel.INFO.Printf("task %d complete! \n", ud.TaskId)
				ud.Complete = true
			}
		}
		if needUpdate {
			revel.INFO.Println("update daily_task:", focus.DailyTask)
			col.UpdateId(focus.Id, bson.M{"$set": bson.M{"daily_task": focus.DailyTask}})
		}
	}
	reward := revel.Config.StringDefault("nv.code", "i-113102-1")
	return c.RenderJson(bson.M{"reward": reward})
}

// 获取奖励
func (c *UserInvCtrl) GetReward(username string, zoneId, taskId int) revel.Result {
	col := models.Col(c.MSession, models.COL_USER_INV)
	user, err := models.FindUserInv(col, username, zoneId)
	if err != nil {
		return c.RenderJson(FailWithMsg("找不到用户"))
	}
	isDaily := true
	baseTask := models.FindBaseDailyTaskId(taskId)
	if baseTask == nil {
		isDaily = false
		baseTask = models.FindBaseMaidTaskId(taskId)
	}
	var ud *models.UserTaskStatus
	if isDaily {
		ud, _ = findUserDailyTaskByTaskId(taskId, user.DailyTask)
	} else {
		ud, _ = findUserDailyTaskByTaskId(taskId, user.MainTask)
	}
	if ud.Got {
		return c.RenderJson(FailWithMsg("您已经领取该奖励了"))
	}
	if !ud.Complete {
		return c.RenderJson(FailWithMsg("您还没有完成该任务"))
	}
	ud.Got = true
	if isDaily {
		revel.INFO.Println("user.DailyTask = ", user.DailyTask)
		col.UpdateId(user.Id, bson.M{"$set": bson.M{"daily_task": user.DailyTask}})
	} else {
		revel.INFO.Println("user.MainTask = ", user.MainTask)
		col.UpdateId(user.Id, bson.M{"$set": bson.M{"main_task": user.MainTask}})
	}
	return c.RenderJson(bson.M{"ok": true, "reward": baseTask.Reward})
}

// 我的被邀请人信息
func (c *UserInvCtrl) FocusList(username string, zoneId int) revel.Result {
	col := models.Col(c.MSession, models.COL_USER_INV)
	user, err := models.FindUserInv(col, username, zoneId)
	if err != nil {
		return c.RenderJson(FailWithMsg("找不到该玩家"))
	}
	all := []*models.UserInvShowInfo{}
	focusUsers := []models.UserInv{}
	col.Find(bson.M{"focus": user.Id}).All(&focusUsers)
	for _, focus := range focusUsers {
		gs := models.FindGameServer(focus.ZoneId)
		url := gs.UserInvInfoUrl(focus.Username)
		info := &models.UserInvShowInfo{}
		info.Name = "未知"
		info.Zone = fmt.Sprintf("%d-%s", gs.ZoneId, gs.Name)
		info.Same = gs.ZoneId == user.ZoneId
		info.ZoneId = gs.ZoneId
		all = append(all, info)
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		json.Unmarshal(data, info)
	}
	rst := bson.M{"list": all}
	if user.Focus.Valid() {
		mine := &models.UserInv{}
		err = col.Find(bson.M{"_id": user.Focus}).One(mine)
		if err == nil {
			gs := models.FindGameServer(mine.ZoneId)
			url := gs.UserInvInfoUrl(mine.Username)
			info := &models.UserInvShowInfo{}
			info.Name = "未知"
			info.Zone = fmt.Sprintf("%d-%s", gs.ZoneId, gs.Name)
			info.Same = gs.ZoneId == user.ZoneId
			info.ZoneId = gs.ZoneId
			rst["mine"] = info
			resp, err := http.Get(url)
			if err == nil {
				defer resp.Body.Close()
				data, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					json.Unmarshal(data, info)
				}
			}
		} else {
			revel.ERROR.Println("err = ", err)
		}
	}
	return c.RenderJson(rst)
}

func findUserDailyTaskByTaskId(taskId int, dailyTasks []*models.UserTaskStatus) (task *models.UserTaskStatus, ok bool) {
	for _, task := range dailyTasks {
		if task.TaskId == taskId {
			return task, true
		}
	}
	return nil, false
}

func taskStatusCompletionSize(arr []*models.UserTaskStatus) int {
	size := 0
	for _, o := range arr {
		if o.Complete {
			size += 1
		}
	}
	return size
}

func userInvMainTaskIsComplete(bm *models.BaseInv, userTaskInfos []*models.UserTaskInfo) bool {
	switch bm.SubType {
	case "LEVEL":
		return bm.Size <= findUserTaskInfoLevelGtSize(userTaskInfos, bm.Condition)
	case "VIP":
		return bm.Size <= findUserTaskInfoVipGtSize(userTaskInfos, bm.Condition)
	default:
		revel.ERROR.Println("unknown bm.SubType = ", bm.SubType)
		return false
	}
}

func findUserTaskInfoLevelGtSize(userTaskInfos []*models.UserTaskInfo, level int) int {
	size := 0
	for _, i := range userTaskInfos {
		revel.INFO.Printf("user level = %d, toLevel = %d \n", i.Level, level)
		if i.Level >= level {
			size += 1
		}
	}
	revel.INFO.Println("size = ", size)
	return size
}

func findUserTaskInfoVipGtSize(userTaskInfos []*models.UserTaskInfo, vip int) int {
	size := 0
	for _, i := range userTaskInfos {
		if i.Vip >= vip {
			size += 1
		}
	}
	return size
}
