package chat

import (
	pb "github.com/golang/protobuf/proto"
	"github.com/revel/revel"
	"github.com/wangboo/wwa/app/chat/proto"
	"github.com/wangboo/wwa/app/models"
	"net"
)

// 跨服聊天协议
type ChatProtocol struct {
}

func (p ChatProtocol) Init() {
}

// 处理消息
func (p ChatProtocol) HandleMsg(zoneId int32, bin []byte, conn net.Conn) (err error) {
	chatMsg := &proto.ChatMsg{}
	err = pb.Unmarshal(bin, chatMsg)
	if err != nil {
		revel.ERROR.Println("unmarshal error", err)
		return
	}
	// 填充 ZoneName 字段
	gs := models.FindGameServer(int(zoneId))
	if gs != nil {
		chatMsg.ZoneName = pb.String(gs.Name)
	}
	// 重新打包
	bin2, err := pb.Marshal(chatMsg)
	if err != nil {
		revel.ERROR.Println("marshal chatMsg error", err)
		return
	}
	// 构造广播消息
	msgSt := brocast_msg{
		zone_id: zoneId,
		msgId:   proto.MessageId_Chat,
		bin:     bin2,
	}
	// 广播出去
	server.cli_brocast_chan <- msgSt
	return
}
