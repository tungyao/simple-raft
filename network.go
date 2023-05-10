package simple_raft

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
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
func (ne *network) ConnectMaster(master *Node) {
	// 正在连接master节点
	if master.Net == nil {
		master.Net = new(network)
	}
	log.Println("正在连接master节点", master.TcpAddr)
	if master.Net.DialConn == nil {
		conn, err := net.Dial("tcp", master.TcpAddr)
		if err != nil {
			log.Fatalln("connect master failed", err)
		}
		master.Net.DialConn = conn
	}

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
	go handleConnection(ne, master.Net.DialConn, func() {
		// 向master节点同步资料
	})

	send = append(send, d...)
	master.Net.DialConn.Write(send)
	// TODO 粘包问题以后在解决

	time.Sleep(time.Second)
	master.Net.DialConn.Write([]byte{0, 0, 4})

}

func (ne *network) Run() {

	// TODO 网络功能需要重新设计
	ln, err := net.Listen("tcp", ne.self.TcpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ln.Close()

	log.Println("Listening on :", ne.self.TcpAddr)
	connections := make(map[net.Conn]context.CancelFunc)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		ctx := context.Background()
		_, connections[conn] = context.WithCancel(ctx)
		go handleConnection(ne, conn, nil)

	}
}

func handleConnection(mainNe *network, conn net.Conn, fun func()) {
	// 正确的需求是什么
	// 节点都是以listen方式监听在某一个端口上，其他节点加入的时候，需要记录conn结构体，还需要同时发送和接受消息
	// 同时接收和发送数据
	var thisId string
	for {
		// 接收数据
		var data = make([]byte, 1024)
		if conn == nil {
			log.Println("准备退出协程")
			return
		}
		n, err := conn.Read(data)
		log.Println(err)
		if err != nil && (err == io.EOF || err == io.ErrUnexpectedEOF) {
			// 连接已经断开了
			log.Println(thisId, "--------------连接已断开")
			mainNe.self.Mux.Lock()
			delete(mainNe.self.allNode, thisId)
			mainNe.self.Mux.Unlock()
			return
		}
		go func([]byte, int) {

			// TODO 这里需要一个数据汇总的东西,什么是数据汇总的东西，其他节点向该节点发送信息
			// 思考 是不是要用channel ,有没有更加实用的方法
			// 加入集群
			log.Println("*******get command", data[2])
			switch data[2] {
			case 1:
				mainNe.self.LastLoseTime = time.Now().UnixMilli()
				log.Println("收到心跳")
				mainNe.self.Status = Follower
				mainNe.self.Channel <- Follower
			case 2: // 接受到新节点的加入 返回 3确认

				nd := &networkNode{}
				log.Println(data, n, string(data[3:n]), data[3:n])
				err = json.Unmarshal(data[3:n], nd)
				if err != nil {
					log.Fatalln(err)
				}
				log.Printf("get new node %+v \n", nd)
				newNode := &Node{
					TcpAddr: nd.TcpAddr,
					RpcAddr: nd.RpcAddr,
					Id:      nd.Id,
					Timeout: nd.Timeout,
					Net:     &network{ListenConn: conn},
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
				// 应该是向其他节点发送新节点的加入通知
				for _, node := range mainNe.self.allNode {
					if node.Id != nd.Id {
						send[2] = 8
						log.Println("向", node.Id, "发送新节点同步请求") // TODO 好像没有生效
						node.Net.ListenConn.Write(send)
					} else {
						send[2] = 3
						log.Println("向", node.Id, "发送新节点请求") // TODO 好像没有生效
						node.Net.ListenConn.Write(send)
					}
				}
				//
			case 3: // 接收到master节点的确认资料 加入到全局节点中
				nd := &networkNode{}
				log.Println(data)
				json.Unmarshal(data[3:n], nd)
				log.Printf("set master  %+v \n", nd)
				newNode := &Node{
					TcpAddr: nd.TcpAddr,
					RpcAddr: nd.RpcAddr,
					Id:      nd.Id,
					Timeout: nd.Timeout,
					Net:     &network{ListenConn: conn, DialConn: conn},
				}
				mainNe.self.Timer.Pause()
				mainNe.self.Mux.Lock()
				thisId = nd.Id
				mainNe.self.allNode[nd.Id] = newNode
				//mainNe.self.MasterName = nd.Id
				mainNe.self.Mux.Unlock()
				mainNe.self.Channel <- Follower
				mainNe.self.Status = Follower
				mainNe.self.Timer.Restart()

				// 向连接节点同步资料
				//mainNe.self.Net.Sync()

			case 4: // 收到其他节点同步节点的信息
				log.Println("收到其他节点同步节点的信息")
				nodes := make([]*networkNode, 0)
				for _, node := range mainNe.self.allNode {
					nodes = append(nodes, &networkNode{
						TcpAddr: node.TcpAddr,
						Id:      node.Id,
						Timeout: node.Timeout,
						RpcAddr: node.RpcAddr,
					})
				}
				syncDataPre := &syncData{}
				syncDataPre.Node = nodes
				d, err := json.Marshal(syncDataPre)
				if err != nil {
					log.Fatalln(err)
				}
				log.Println(string(d))
				conn.Write(append([]byte{0, 0, 5}, d...))
			case 5:
				syncDataPre := &syncData{}
				err = json.Unmarshal(data[3:n], syncDataPre)
				if err != nil {
					log.Fatalln(err)
				}
				log.Println(data)
				for _, node := range syncDataPre.Node {
					log.Println(node)
					if node.Id != mainNe.self.Id {
						mainNe.self.allNode[node.Id] = &Node{
							TcpAddr: node.TcpAddr,
							Timeout: node.Timeout,
							Id:      node.Id,
							Net:     new(network),
							Vote:    new(Vote),
						}
					}
				}
				log.Println("同步完成")
			case 6:
			case 7: // 与master节点建立连接
				mainNe.self.Mux.RLock()
				mainNe.self.Timer.Pause()
				log.Println(mainNe.self.allNode, string(data[3:n]))
				mainNe.self.MasterName = string(data[3:n])
				if v, ok := mainNe.self.allNode[string(data[3:n])]; ok {
					mainNe.self.Net.ConnectMaster(v)
				}
				mainNe.self.Timer.Restart()
				mainNe.self.Mux.RUnlock()
				mainNe.self.Status = Follower
				mainNe.self.Channel <- Follower
			case 8: // 收到master通知 需要同步新节点
				//mainNe.self.Net.DialConn
				if v, ok := mainNe.self.allNode[mainNe.self.MasterName]; ok {
					v.Net.DialConn.Write([]byte{0, 0, 4})
				}
			}
		}(data, n)
	}

}

// HeartRequest 向其他节点发送心跳 以超时事件最短的为发送时间
func (ne *network) HeartRequest() {
	// TODO send heart to each node
	log.Println(ne.self.Id, "向其他发送心跳")
	log.Println(ne.self.allNode)
	ne.self.Mux.RLock()
	defer ne.self.Mux.RUnlock()
	for _, v := range ne.self.allNode {
		if v.Net.ListenConn != nil {
			log.Println(v.Id, "send heart")
			v.Net.ListenConn.Write([]byte{0, 0, 1})
		} else {
			ne.self.Net.MasterConnect(v.TcpAddr)
		}
	}
}

// MasterConnect CheckMasterCheckSlaveConnect 与其他子节点建立连接
func (ne *network) MasterConnect(tcpAddr string) {
	send := []byte{0, 0, 7}
	dial, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		log.Fatalln("发送失败", err)
	}
	send = append(send, []byte(ne.self.Id)...)
	_, err = dial.Write(send)
	dial.Close()
}

// NoticeAnotherNodeJoined 通知其他节点新节点的加入
func (ne *network) NoticeAnotherNodeJoined() {

}

type syncData struct {
	Node []*networkNode `json:"node"`
}

func (ne *network) Sync() {
	_, err := ne.DialConn.Write([]byte{0, 0, 4})
	if err != nil {
		log.Println(err)
		return
	}
}

// Pack 打包
func (ne *network) Pack(data []byte) {
	for i := len(data) / (1024 - 2 - 3); i < len(data); i++ {

	}
}

// Unpack 解包
func (ne *network) Unpack() {

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

func uint642uint8[T int | uint64, S uint8](n T, arr []S) []S {
	arr = append(arr,
		S(n&0xff),
		S((n>>8)&0xff),
		S((n>>16)&0xff),
		S((n>>24)&0xff),
		S((n>>32)&0xff),
		S((n>>40)&0xff),
		S((n>>48)&0xff),
		S((n>>56)&0xff),
	)
	return arr
}
