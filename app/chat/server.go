package chat

import (
	"bufio"
	"fmt"
	"github.com/revel/revel"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	server   *chat_server
	start_at time.Time
)

// 服务器
type chat_server struct {
	cli_map       map[int]*chat_client // 连接的客户端集合
	cli_quit_chan chan int             // 客户端退出服务器
	cli_join_chan chan *chat_client    // 客户端加入服务器
	cli_chat_chan chan chat_msg        // 客户端广播聊天消息
	send_times    int                  // 发送次数
	msg_count     int                  // 消息数
}

// 连接客户端信息
type chat_client struct {
	zone_id int      // 服务器序号
	conn    net.Conn // 客户端连接
}

// 客户端广播聊天消息
type chat_msg struct {
	zone_id int
	msg     string
}

// addr = ":10002"
// 启动服务器，该方法不会阻塞
func Start(addr string) {
	server = &chat_server{
		cli_map:       map[int]*chat_client{},
		cli_quit_chan: make(chan int),
		cli_join_chan: make(chan *chat_client),
		cli_chat_chan: make(chan chat_msg, 1024), // 聊天通道有1024个缓存空间，防止阻塞客户端进程
		send_times:    0,
		msg_count:     0,
	}
	start_at = time.Now()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	go accept(ln)
	go server_manage_process()
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
		case msgSt := <-server.cli_chat_chan:
			// 广播聊天消息
			for zone_id, cli := range server.cli_map {
				if msgSt.zone_id != zone_id {
					// 广播给其他客户端
					cli.conn.Write([]byte(msgSt.msg))
					server.send_times += 1
				}
			}
			server.msg_count += 1
		}
	}
}

// 处理客户端连接
func process_socket(conn net.Conn) {
	reader := bufio.NewReader(conn)
	// 握手
	zoneIdStr, err := reader.ReadString('\n')
	if err != nil {
		revel.ERROR.Println("shakeHand fail", err)
		return
	}
	// 解析客户端编号
	zoneId, err := strconv.Atoi(strings.TrimSpace(zoneIdStr))
	if err != nil {
		revel.ERROR.Printf("shankeHand fail when parse zone_id(%s), error: %v\n", zoneIdStr, err)
	}
	fmt.Printf("客户端%d号已连接\n", zoneId)
	// 加入服务器进程
	cliSt := &chat_client{
		zone_id: zoneId,
		conn:    conn,
	}
	server.cli_join_chan <- cliSt
	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			// 读取socket报错，断开连接
			revel.ERROR.Printf("reading conn error zoneId(%d), error: %v\n", zoneId, err)
			server.cli_quit_chan <- zoneId
			conn.Close()
			break
		}
		// 发送空消息过来为心跳包
		if str != "\n" {
			// fmt.Printf("recv from %d: %s", zoneId, str)
			// 将消息打包给服务进程广播出去
			msgSt := chat_msg{
				zone_id: zoneId,
				msg:     str,
			}
			server.cli_chat_chan <- msgSt
		}
	}
}

func Report() {
	cost := time.Now().Unix() - start_at.Unix()
	avg := float32(server.send_times) / float32(cost)
	fmt.Printf("总消息数:%d, 分发消息次数:%d, 平均每秒分发次数%f\n", server.msg_count, server.send_times, avg)
}
