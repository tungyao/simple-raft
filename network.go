package simple_raft

import (
	"fmt"
	"net"
	"os"
	"unsafe"
)

// 换一种思路
// 在本地维护多个角色 当时node移交到哪个角色上时 就使用那些方法

// 网络相关的东西
type network struct {
	self    *Node
	Address string
	Rece    chan []byte
	Send    chan []byte
	Conn    net.Conn
}

// In fact , there's a node list in program
// so, Reload from the yaml file

// 协议 按照byte=8位
// 1 | 2          | 3                  | 4 | 5               | 6 | 7 | 8 |
//  是否接续上一报文   操作码
//                   0 get heart
//                   1 send heart
//                   2 join group        heart for timeout
//

func (ne *network) Run() {

	// TODO 网络功能需要重新设计
	ne.Rece = make(chan []byte, 512)
	ne.Send = make(chan []byte, 512)
	ln, err := net.Listen("tcp", ne.self.Addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ln.Close()

	fmt.Println("Listening on :", ne.self.Addr)
	connections := make(map[net.Conn]bool)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		connections[conn] = true
		go handleConnection(conn, connections)

	}
}

func handleConnection(conn net.Conn, connections map[net.Conn]bool) {
	// TODO 这里是写个狗屁
	// 正确的需求是什么
	// 节点都是以listen方式监听在某一个端口上，其他节点加入的时候，需要记录conn结构体，还需要同时发送和接受消息
	// 同时接收和发送数据
	var data [1024]byte
	for {
		// 接收数据
		_, err := conn.Read(data[:])
		if err != nil {
			return
		}
		// TODO 这里需要一个数据汇总的东西

		// 加入集群
		if data[2] == 0x2 {
			// 声明超时时间
			nd := &Node{
				Timeout: int(uint32(data[4]&0xff)<<8) + int(data[3]&0xff),
				Net:     &network{Conn: conn},
			}
			mux.Lock()
			allNode = append(allNode, nd)
			mux.Unlock()
		}

	}

}

func sendData(conn net.Conn) {
	for {
		fmt.Print("Enter text: ")
		var input string
		fmt.Scanln(&input)

		if input == "/quit" {
			conn.Close()
			os.Exit(0)
		}

		_, err := conn.Write([]byte(input))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func receiveData(conn net.Conn) {
	buffer := make([]byte, 1024)

	for {
		size, err := conn.Read(buffer)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		data := string(buffer[:size])
		fmt.Println("Received:", data)
	}
}

// HeartRequest 向其他节点发送心跳 以超时事件最短的为发送时间
func (ne *network) HeartRequest(nodes []*Node) {
	// TODO send heart to each node
	for _, v := range allNode {
		v.Net.Send <- []byte{0, 0, 1}
	}
}

func uint82Uint64[T int | uint8](r ...T) uint64 {
	return uint64(r[0]&0xff) +
		uint64(r[1]&0xff)<<8 +
		uint64(r[2]&0xff)<<16 +
		uint64(r[3]&0xff)<<24 +
		uint64(r[4]&0xff)<<32 +
		uint64(r[5]&0xff)<<40 +
		uint64(r[6]&0xff)<<48 +
		uint64(r[7]&0xff)<<56
}

func uint642uint8[T int | uint8 | uint64](n T, arr unsafe.Pointer) {
	var arr1 = (*[]T)(arr)
	*arr1 = append(*arr1,
		T(n&0xff),
		T((n>>8)&0xff),
		T((n>>16)&0xff),
		T((n>>24)&0xff),
		T((n>>32)&0xff),
		T((n>>40)&0xff),
		T((n>>48)&0xff),
		T((n>>56)&0xff),
	)
}
