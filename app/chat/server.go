package chat

import (
	"errors"
	"fmt"
	pb "github.com/golang/protobuf/proto"
	"github.com/revel/revel"
	"github.com/wangboo/wwa/app/chat/proto"
	"io"
	"net"
	// "strconv"
	// "strings"
	"time"
)

var (
	server   *chat_server
	start_at time.Time
)

// 需实现协议
type Protocol interface {
	Init()
	HandleMsg(zoneId int32, bin []byte, conn net.Conn) (err error)
}

// 处理协议
// func (p *Protocol) HandleMsg(zoneId int32, bin []byte) {

// }

// 服务器
type chat_server struct {
	protocol_map     map[proto.MessageId]*Protocol // 协议处理器集合
	cli_map          map[int32]*chat_client        // 连接的客户端集合
	cli_quit_chan    chan int32                    // 客户端退出服务器
	cli_join_chan    chan *chat_client             // 客户端加入服务器
	cli_brocast_chan chan brocast_msg              // 客户端广播聊天消息
	send_times       int                           // 发送次数
	msg_count        int                           // 消息数
}

// 连接客户端信息
type chat_client struct {
	zone_id int32    // 服务器序号
	conn    net.Conn // 客户端连接
}

// 客户端广播聊天消息
type brocast_msg struct {
	zone_id int32           // 过滤服务器
	msgId   proto.MessageId // 消息id
	bin     []byte          // 消息内容
}

// addr = ":10002"
// 启动服务器，该方法不会阻塞
func Start(addr string) {
	server = &chat_server{
		protocol_map:     map[proto.MessageId]*Protocol{},
		cli_map:          map[int32]*chat_client{},
		cli_quit_chan:    make(chan int32),
		cli_join_chan:    make(chan *chat_client),
		cli_brocast_chan: make(chan brocast_msg, 1024), // 聊天通道有1024个缓存空间，防止阻塞客户端进程
		send_times:       0,
		msg_count:        0,
	}
	start_at = time.Now()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	go accept(ln)
	go server_manage_process()
}

// 添加一个协议处理器
func AddProtocol(msgId proto.MessageId, protocol Protocol) {
	protocol.Init()
	server.protocol_map[msgId] = &protocol
}

// 侦听端口，该方法不会退出
func accept(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			revel.ERROR.Println("accept error", err)
			continue
		}
		fmt.Println("有新客户端连接")
		// 握手
		go process_socket(conn)
	}
}

// 服务器广播和客户端管理进程
func server_manage_process() {
	for {
		select {
		case zone_id := <-server.cli_quit_chan:
			// 客户端进程退出信号量
			fmt.Printf("client %d exit\n", zone_id)
			delete(server.cli_map, zone_id)
		case msgSt := <-server.cli_join_chan:
			// 客户端加入信号
			fmt.Printf("client %d join\n", msgSt.zone_id)
			server.cli_map[msgSt.zone_id] = msgSt
		case msgSt := <-server.cli_brocast_chan:
			// 广播聊天消息
			for zone_id, cli := range server.cli_map {
				if msgSt.zone_id != zone_id {
					// 广播给其他客户端
					write_message(cli.conn, msgSt.msgId, msgSt.bin)
					server.send_times += 1
				}
			}
			server.msg_count += 1
		}
	}
}

// 写消息
func write_message(conn net.Conn, msgid proto.MessageId, bin []byte) (err error) {
	_, err = conn.Write(int32_to_byte4(int32(msgid)))
	if err != nil {
		return
	}
	_, err = conn.Write(int32_to_byte4(int32(len(bin))))
	if err != nil {
		return
	}
	_, err = conn.Write(bin)
	return
}

func int32_to_byte4(v int32) []byte {
	byteV := make([]byte, 4)
	byteV[0] = byte((v >> 24) & 0xff)
	byteV[1] = byte((v >> 16) & 0xff)
	byteV[2] = byte((v >> 8) & 0xff)
	byteV[3] = byte(v & 0xff)
	return byteV
}

// 处理客户端连接
func process_socket(conn net.Conn) {
	// 读取握手信息
	var err error
	var zoneId int32
	// 读取二进制数据
	var bin []byte
	var msgId proto.MessageId
	for {
		msgId, bin, err = readPkg(conn)
		if err != nil {
			// 读取socket报错，断开连接
			revel.ERROR.Printf("reading conn error zoneId(%d), error: %v\n", zoneId, err)
			break
		}
		fmt.Printf("recv client: ZoneId %d, msgId %v, bin %v \n", zoneId, msgId, bin)
		// 心跳
		if msgId == proto.MessageId_HeartBeat {
			continue
		} else if msgId == proto.MessageId_Login {
			// 登陆报文
			login := &proto.Login{}
			err = pb.Unmarshal(bin, login)
			if err != nil {
				break
			}
			zoneId = login.GetZoneId()
			fmt.Printf("客户端%d号已连接\n", zoneId)
			// 加入服务器进程
			cliSt := &chat_client{
				zone_id: zoneId,
				conn:    conn,
			}
			server.cli_join_chan <- cliSt
			continue
		}
		err := process_pkg(zoneId, msgId, bin, conn)
		if err != nil {
			fmt.Println("process_socket err", err)
			break
		}
	}
	server.cli_quit_chan <- zoneId
	conn.Close()
}

// 处理消息包
func process_pkg(zoneId int32, msgId proto.MessageId, bin []byte, conn net.Conn) (err error) {
	// 容错处理
	defer func() {
		if r := recover(); r != nil {
			revel.ERROR.Println("process pkg fail,", r)
		}
	}()
	// 处理消息
	protocol, ok := server.protocol_map[msgId]
	if ok {
		err = (*protocol).HandleMsg(zoneId, bin, conn)
	} else {
		revel.ERROR.Println("process pkg recv unknown msg: ", msgId)
	}
	return
}

// 读取一个消息包
func readPkg(conn net.Conn) (msgId proto.MessageId, bin []byte, err error) {
	msgIdInt, err := readInt32(conn)
	if err != nil {
		return
	}
	fmt.Println("readPkg msgIdInt: ", msgIdInt)
	msgId = proto.MessageId(msgIdInt)
	if msgId == proto.MessageId_HeartBeat {
		// 心跳报文，没有包长度和包体
		return
	}
	// 读取包长度
	binLen, err := readInt32(conn)
	if err != nil {
		return
	}
	fmt.Println("readPkg binLen: ", binLen)
	// 读取包内容
	bin = make([]byte, binLen)
	_, err = io.ReadFull(conn, bin)
	fmt.Println("readPkg bin: ", bin)
	return
}

// 读取下一个int32(4byte)
func readInt32(conn net.Conn) (len int32, err error) {
	intBytes := make([]byte, 4)
	_, err = io.ReadFull(conn, intBytes)
	if err != nil {
		// revel.ERROR.Println("shakeHand fail when read msgId", err)
		return -1, err
	}
	return byte4ToInt32(intBytes)
}

// 4个byte转int32
func byte4ToInt32(bin []byte) (int32, error) {
	if len(bin) != 4 {
		return 0, errors.New("bin is not 4 byte")
	}
	return (int32(bin[0]) << 24) +
		(int32(bin[1]) << 16) +
		(int32(bin[2]) << 8) +
		int32(bin[3]), nil
}

func Report() {
	cost := time.Now().Unix() - start_at.Unix()
	avg := float32(server.send_times) / float32(cost)
	fmt.Printf("总消息数:%d, 分发消息次数:%d, 平均每秒分发次数%f\n", server.msg_count, server.send_times, avg)
}
