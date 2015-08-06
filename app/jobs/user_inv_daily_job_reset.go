package mjob

import (
	"github.com/wangboo/wwa/app/models"
	"labix.org/v2/mgo/bson"
)

// 获取每日竞技人员
type UserInvDailyJobReset struct {
}

func (r *UserInvDailyJobReset) Run() {
	defer catchException()
	r.RunImpl()
}

func (r *UserInvDailyJobReset) RunImpl() {
	s := models.Session()
	defer s.Close()
	c := s.DB(models.DB_NAME).C(models.COL_USER_INV)
	DailyTasks := []*models.UserTaskStatus{}
	for _, bd := range models.BaseDailyInvList {
		userDaily := &models.UserTaskStatus{bd.Id, []bson.ObjectId{}, false, false}
		DailyTasks = append(DailyTasks, userDaily)
	}
	c.UpdateAll(bson.M{}, bson.M{"$set": bson.M{"daily_task": DailyTasks}})
}
