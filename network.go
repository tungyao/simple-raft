package simple_raft

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
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

func handleConnection(mainNe *network, conn net.Conn, fun func()) {
	// 正确的需求是什么
	// 节点都是以listen方式监听在某一个端口上，其他节点加入的时候，需要记录conn结构体，还需要同时发送和接受消息
	// 同时接收和发送数据
	var thisId string
	for {
		// 接收数据
		if conn == nil {
			return
		}
		data, err := receive(conn)
		if err != nil && (err == io.EOF || err == io.ErrUnexpectedEOF) {
			// 连接已经断开了
			mainNe.self.Mux.Lock()
			if thisId == mainNe.self.MasterName {
				mainNe.self.MasterName = ""
			}
			delete(mainNe.self.allNode, thisId)
			mainNe.self.Mux.Unlock()
			return
		}
		if len(data) == 0 {
			return
		}
		if len(data) > 2 {
			log.Println("get command", data[2])
		}
		go func([]byte) {

			// 加入集群
			switch data[2] {
			case 1:
				mainNe.self.LastLoseTime = time.Now().UnixMilli()
				mainNe.self.Status = Follower
				mainNe.self.Channel <- Follower
			case 2: // 接受到新节点的加入 返回 3确认

				nd := &networkNode{}
				err = json.Unmarshal(data[3:], nd)
				if err != nil {
				}
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
				}
				send = append(send, d...)
				mainNe.self.Mux.Unlock()
				// 应该是向其他节点发送新节点的加入通知
				for _, node := range mainNe.self.allNode {
					if node.Id != nd.Id {
						send[2] = 8
						node.Net.ListenConn.Write(pack(send))
					} else {
						send[2] = 3
						node.Net.ListenConn.Write(pack(send))
					}
				}
				//
			case 3: // 接收到master节点的确认资料 加入到全局节点中
				nd := &networkNode{}
				json.Unmarshal(data[3:], nd)
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
				}
				conn.Write(pack(append([]byte{0, 0, 5}, d...)))
			case 5: // 收到master节点同步
				syncDataPre := &syncData{}
				err = json.Unmarshal(data[3:], syncDataPre)
				if err != nil {
				}
				for _, node := range syncDataPre.Node {
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
			case 6:
				otherTerm := uint82Uint64(data[3], data[4], data[5], data[6], data[7], data[8], data[9], data[10])
				otherIndex := uint82Uint64(data[11], data[12], data[13], data[14], data[15], data[16], data[17], data[18])
				send := []byte{0, 0, 7}
				log.Println("收到请求票 other term", otherTerm, "other log", otherIndex, "self term", mainNe.self.TermIndex, "self log", mainNe.self.LogIndex)
				if mainNe.self.TermIndex > otherTerm {
					// term index 更大，优先级更高
					send = append(send, 0)

				} else if mainNe.self.TermIndex < otherTerm {
					// term index 更小，优先级更低
					send = append(send, 1)

				} else {
					// term index 相同，比较 log index
					if mainNe.self.LogIndex > otherIndex {
						// log index 更大，优先级更高
						send = append(send, 0)

					} else {
						// log index 更小或相等，优先级更低
						send = append(send, 1)
					}
				}
				log.Println("发送投票", send[3])
				conn.Write(pack(send))
			case 7: // 与master节点建立连接
				mainNe.self.Mux.RLock()
				mainNe.self.Timer.Pause()
				mainNe.self.MasterName = string(data[3:])
				if v, ok := mainNe.self.allNode[string(data[3:])]; ok {
					mainNe.self.Net.ConnectMaster(v)
				}
				mainNe.self.Timer.Restart()
				mainNe.self.Mux.RUnlock()
				mainNe.self.Status = Follower
				mainNe.self.Channel <- Follower
			case 8: // 收到master通知 需要同步新节点
				//mainNe.self.Net.DialConn
				if v, ok := mainNe.self.allNode[mainNe.self.MasterName]; ok {
					v.Net.DialConn.Write(pack([]byte{0, 0, 4}))
				}
			case 9: // 接受到外部的信息
				atomic.AddUint64(&mainNe.self.LogIndex, 1)
				mainNe.self.Message <- data[8:]
			case 10: // 接收到master的复制日志
				logIndex := uint82Uint64(data[3], data[4], data[5], data[6], data[7], data[8], data[9], data[10])
				mainNe.self.Data[logIndex] = data[11:]
				send := []byte{0, 0, 11}
				send = append(send, data[3], data[4], data[5], data[6], data[7], data[8], data[9], data[10])
				conn.Write(pack(send))
				log.Println("接受到master日志", string(data[11:]))
			case 11: // 收到节点反馈
				logIndex := uint82Uint64(data[3], data[4], data[5], data[6], data[7], data[8], data[9], data[10])
				// 计算和判断写否要写入持久存储
				if v, ok := mainNe.self.Buff[logIndex]; ok && v.vote > len(mainNe.self.allNode) {
					mainNe.self.Data[logIndex] = v.data
				}
			}

		}(data)
	}

}

func (ne *network) Broadcast(data *[]byte) {
	n := atomic.AddUint64(&ne.self.LogIndex, 1)
	da := []byte{0, 0, 10}
	da = uint642uint8(n, da)
	da = append(da, *data...)
	// 提交到缓存区
	ne.self.Buff[n] = &BuffData{
		data: *data,
		vote: 0,
	}
	for _, node := range ne.self.allNode {
		node.Net.ListenConn.Write(pack(da))
	}
}

// HeartRequest 向其他节点发送心跳 以超时事件最短的为发送时间
func (ne *network) HeartRequest() {
	// TODO send heart to each node
	ne.self.Mux.RLock()
	defer ne.self.Mux.RUnlock()
	for _, v := range ne.self.allNode {
		if v.Net.ListenConn != nil {
			v.Net.ListenConn.Write(pack([]byte{0, 0, 1}))
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
	}
	send = append(send, []byte(ne.self.Id)...)
	_, err = dial.Write(pack(send))
	dial.Close()
}

// ConnectMaster TODO 需要有一个先master节点主动连接的动作
func (ne *network) ConnectMaster(master *Node) {
	// 正在连接master节点
	if master.Net == nil {
		master.Net = new(network)
	}
	if master.Net.DialConn == nil {
		conn, err := net.Dial("tcp", master.TcpAddr)
		if err != nil {
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
	}
	go handleConnection(ne, master.Net.DialConn, nil)

	send = append(send, d...)
	master.Net.DialConn.Write(pack(send))
	// TODO 粘包问题以后在解决

	time.Sleep(time.Second)
	master.Net.DialConn.Write(pack([]byte{0, 0, 4}))

}

func (ne *network) Run() {

	// TODO 网络功能需要重新设计
	ln, err := net.Listen("tcp", ne.self.TcpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go handleConnection(ne, conn, nil)

	}
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
		return
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

// --------
// 解决tcp粘包的问题

const (
	PacketHeadLength = 4 // 数据包头部长度为 7
)

// Pack 打包
func pack(data []byte) []byte {
	size := len(data)
	buffer := make([]byte, size+PacketHeadLength)
	binary.BigEndian.PutUint32(buffer[:PacketHeadLength], uint32(size))
	copy(buffer[PacketHeadLength:], data)
	return buffer
}

// Unpack 解包
func unpack(buffer []byte) ([]byte, uint32, error) {
	size := binary.BigEndian.Uint32(buffer[:PacketHeadLength])
	if uint32(len(buffer)) < (size + PacketHeadLength) {
		return nil, 0, fmt.Errorf("packet too small")
	}
	return buffer[PacketHeadLength : PacketHeadLength+size], size, nil
}
func receive(conn net.Conn) ([]byte, error) {
	buffer := make([]byte, 1024)
	result := make([]byte, 0)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return nil, err
		}
		data, size, err := unpack(buffer[:n])
		if err != nil {
			return nil, err
		}
		log.Println(size)
		result = append(result, data...)
		if int(size) == len(result) {
			break
		}
	}
	return result, nil
}
