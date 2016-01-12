package chat

import (
	"fmt"
	pb "github.com/golang/protobuf/proto"
	"net"
	// "github.com/revel/revel"
	"github.com/wangboo/wwa/app/chat/proto"
	"github.com/wangboo/wwa/app/models"
)

var (
	behaviourJobChan chan behaviourData
)

// 行为处理消息结构
type behaviourData struct {
	beh    proto.Behaviour
	zoneId int32
}

type BehaviourProtocol struct {
}

func (p BehaviourProtocol) Init() {
	behaviourJobChan = make(chan behaviourData, 2048)
	go behaviourJobProcessing()
}

// 处理消息
func (p BehaviourProtocol) HandleMsg(zoneId int32, bin []byte, conn net.Conn) (err error) {
	beh := &proto.Behaviour{}
	err = pb.Unmarshal(bin, beh)
	if err != nil {
		return
	}
	// 消息压入处理进程
	fmt.Println("BehaviourProtocol HandleMsg push message\n", beh)
	fmt.Println("zoneId = ", zoneId)
	behaviourJobChan <- behaviourData{beh: *beh, zoneId: zoneId}
	return
}

// 新的处理进程，处理行为日志记录
func behaviourJobProcessing() {
	fmt.Println("behaviourJobProcessing running")
	for {
		select {
		case data := <-behaviourJobChan:
			handleBehaviour(&data)
		}
	}
}

func handleBehaviour(data *behaviourData) {
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		revel.ERROR.Println("handleBehaviour err", err)
	// 	}
	// }()
	fmt.Printf("handleBehaviour, zoneId = %d, data.beh = %v \n", data.zoneId, data.beh.Type)
	switch *(data.beh.Type) {
	case proto.BehaviourType_CreateRole:
		createUser(data.zoneId, data.beh.GetCreateRole())
	case proto.BehaviourType_Guide:
		guideChange(data.zoneId, data.beh.GetGuide())
	case proto.BehaviourType_View:
		viewChange(data.zoneId, data.beh.GetView())
	case proto.BehaviourType_Exp:
		expChange(data.zoneId, data.beh.GetExp())
	case proto.BehaviourType_Battle:
		battleChange(data.zoneId, data.beh.GetBattle())
	}
}

// 创建角色
func createUser(zoneId int32, role *proto.BehaviourCreateRole) {
	models.NewBehaviour(int(zoneId), int(role.GetUserId()), role.GetName(), role.GetPlatform())
}

func guideChange(zoneId int32, guide *proto.BehaviourGuide) {
	models.GuideChange(int(zoneId), int(guide.GetUserId()), int(guide.GetGuideId()))
}

func viewChange(zoneId int32, view *proto.BehaviourView) {
	models.ViewChange(int(zoneId), int(view.GetUserId()), int(view.GetViewId()))
}

func expChange(zoneId int32, exp *proto.BehaviourExp) {
	models.ExpChange(int(zoneId), int(exp.GetUserId()), int(exp.GetLevel()), int(exp.GetExp()))
}

func battleChange(zoneId int32, battle *proto.BehaviourBattle) {
	models.BattleChange(int(zoneId), int(battle.GetUserId()), int(battle.GetBattleId()))
}
