package simple_raft

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
	"unsafe"
)

type networkNode struct {
	TcpAddr string `yaml:"tcp_addr" json:"tcp_addr"`
	Id      string `yaml:"id" json:"id"`
	Timeout int    `json:"timeout"`
	RpcAddr string `yaml:"rpc_addr" json:"rpc_addr"`
}

// 换一种思路
// 在本地维护多个角色 当时node移交到哪个角色上时 就使用那些方法

// 网络相关的东西
type network struct {
	self       *Node
	Address    string
	Rece       chan *networkConn
	Send       chan []byte
	ListenConn net.Conn
	DialConn   net.Conn
}

type networkConn struct {
	Conn        net.Conn
	IsClose     bool
	Data        [1024]byte
	Stop        int
	networkNode *networkNode
	sync.RWMutex
}

// In fact , there's a node list in program
// so, Reload from the yaml file

// 协议 按照byte=8位
// 1 | 2          | 3                  | 4 | 5               | 6 | 7 | 8 |
//
//	是否接续上一报文   操作码
//	                 0 get heart
//	                 1 send heart
//	                 2 join group        heart for timeout
//

// ConnectMaster TODO 需要有一个先master节点主动连接的动作
func (ne *network) ConnectMaster(masterAddress string) {
	// 正在连接master节点
	log.Println("正在连接master节点", masterAddress)
	conn, err := net.Dial("tcp", masterAddress)
	if err != nil {
		log.Fatalln("connect master failed", err)
	}
	ne.DialConn = conn
	send := make([]byte, 0)
	send = append(send, 0, 0, 2)
	d, err := json.Marshal(&networkNode{
		TcpAddr: ne.self.TcpAddr,
		Id:      ne.self.Id,
		Timeout: 3000,
		RpcAddr: ne.self.RpcAddr,
	})
	if err != nil {
		log.Fatalln("connect master failed and json marshal", err)
	}
	send = append(send, d...)
	ne.DialConn.Write(send)

	go handleConnection(ne, conn, nil)
}

func (ne *network) Run() {

	// TODO 网络功能需要重新设计
	ne.Rece = make(chan *networkConn, 512)
	ln, err := net.Listen("tcp", ne.self.TcpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ln.Close()

	log.Println("Listening on :", ne.self.TcpAddr)
	connections := make(map[net.Conn]bool)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		connections[conn] = true
		go handleConnection(ne, conn, connections)

	}
}

func handleConnection(mainNe *network, conn net.Conn, connections map[net.Conn]bool) {
	// 正确的需求是什么
	// 节点都是以listen方式监听在某一个端口上，其他节点加入的时候，需要记录conn结构体，还需要同时发送和接受消息
	// 同时接收和发送数据
	var data [1024]byte
	var thisId string
	for {
		// 接收数据
		n, err := conn.Read(data[:])
		if err != nil && err == io.EOF {
			// 连接已经断开了
			mainNe.self.Mux.Lock()
			if thisId == mainNe.self.MasterName {
				mainNe.self.MasterName = "" // master节点置空
			}
			delete(mainNe.self.allNode, thisId)
			mainNe.self.Mux.Unlock()
			return
		}
		// TODO 这里需要一个数据汇总的东西,什么是数据汇总的东西，其他节点向该节点发送信息
		// 思考 是不是要用channel ,有没有更加实用的方法
		// 加入集群
		log.Println("get command", data[2])
		switch data[2] {
		case 1:
			mainNe.self.LastLoseTime = time.Now().UnixMilli()
			log.Println("收到心跳")
			mainNe.self.Status = Follower
			mainNe.self.Channel <- Follower
		case 2: // 接受到新节点的加入 返回 3确认

			nd := &networkNode{}
			json.Unmarshal(data[3:n], nd)
			log.Printf("get new node %+v \n", nd)
			newNode := &Node{
				TcpAddr: nd.TcpAddr,
				RpcAddr: nd.RpcAddr,
				Id:      nd.Id,
				Timeout: nd.Timeout,
				Net:     &network{ListenConn: conn, DialConn: conn},
			}
			mainNe.self.Mux.Lock()
			thisId = nd.Id
			mainNe.self.allNode[nd.Id] = newNode

			send := make([]byte, 0)
			send = append(send, 0, 0, 3)
			d, err := json.Marshal(&networkNode{
				TcpAddr: mainNe.self.TcpAddr,
				Id:      mainNe.self.Id,
				Timeout: mainNe.self.Timeout,
				RpcAddr: mainNe.self.RpcAddr,
			})
			if err != nil {
				log.Fatalln("connect master failed and json marshal", err)
			}
			send = append(send, d...)
			mainNe.self.Mux.Unlock()
			//conn.Write(send)
			for _, node := range mainNe.self.allNode {
				log.Println("向", node.Id, "发送新节点请求")
				node.Net.ListenConn.Write(send)
			}
			//
		case 3: // 接收到master节点的确认资料 加入到全局节点中
			nd := &networkNode{}
			json.Unmarshal(data[3:n], nd)
			log.Printf("set master  %+v \n", nd)
			newNode := &Node{
				TcpAddr: nd.TcpAddr,
				RpcAddr: nd.RpcAddr,
				Id:      nd.Id,
				Timeout: nd.Timeout,
				Net:     &network{ListenConn: conn},
			}
			mainNe.self.Mux.Lock()
			mainNe.self.Timer.Pause()
			thisId = nd.Id
			mainNe.self.allNode[nd.Id] = newNode
			mainNe.self.MasterName = nd.Id
			mainNe.self.Timer.Restart()
			mainNe.self.Mux.Unlock()

			// 同步master节点的其他节点资料
			FastRpcClientInit(mainNe.self, mainNe.self.allNode[mainNe.self.MasterName].RpcAddr)
		case 4: // 来自master的日志

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
func (ne *network) HeartRequest() {
	// TODO send heart to each node
	log.Println(ne.self.Id, "向其他发送心跳")
	log.Println(ne.self.allNode)
	for _, v := range ne.self.allNode {
		log.Println(v.Id, "send heart")
		if v.Net.DialConn == nil {
			log.Println(v.Id, "建立新的连接")
			conn, err := net.Dial("tcp", v.TcpAddr)
			if err != nil {
				log.Println("与子节点建立失败", err)
				continue
			}
			v.Net.DialConn = conn
		}
		v.Net.DialConn.Write([]byte{0, 0, 1})
	}
}

// CheckSlaveConnect 与其他子节点建立连接
func (ne *network) CheckSlaveConnect() {
	for _, i2 := range ne.self.allNode {
		if i2.Net.DialConn == nil {
			conn, err := net.Dial("tcp", i2.TcpAddr)
			if err != nil {
				log.Println("与子节点建立失败", err)
				continue
			}
			i2.Net.DialConn = conn
		}
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
