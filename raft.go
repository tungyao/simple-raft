// go:doc
// raft算法包括
// 选举 复制 成员变动 日志压缩
// 多个关键词
//  1. 复制状态机 (Replicated state machines) , 只有收到确定指定才会在指定内存地址修改为确定的值,exp: x<-3 => x<-5 ,只有领导接受客户端指令
//  2. 任期 (term),可以看作一个领导人持续的时间段，为一个任期,并且一个无符号自增整型id,称为任期号
//  3. 日志提交 (commit),leader在收到客户端指令时，不会立即执行，而是会产生一个 日志项(log entry),其中包含指令，当log entry被复制到超过一个的节点时,日志项会被leader commit，并执行该项
//
// 3种角色
//  1. leader 主要是复制日志到其他节点,维持心跳(心跳时间要小于全部节点的超市时间)
//  2. candidate
//  3. follower
//
// 领导人选举(Leader Election)

package simple_raft

import (
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"sync"
	"time"
)

// 实现的要求
// 简单和易于阅读 去除不必要的方法 接口

type Raft struct {
	addr []*Node // rpc通讯地址
	id   string  // 主机标识
}

const (
	Follower = iota
	Candidate
	Leader
)

// Node 部署每一个主机
type Node struct {
	TcpAddr      string `yaml:"tcp_addr" json:"tcp_addr"` // 主机地址
	Rate         uint8  // 投票倍率 这个值一般可以忽略
	Id           string `yaml:"id" json:"id"` // 主机标识
	Status       int
	Net          *network // 网络相关的操作
	Timeout      int      `yaml:"timeout" json:"timeout"` // 心跳间隔
	LastLoseTime int64    // 上次收到心跳时间
	Channel      chan int
	Message      chan []byte
	LogIndex     uint64
	TermIndex    uint64
	IsVote       bool // 在当前任期是不是已经投了票
	Timer        *timer
	Mux          sync.RWMutex
	RpcAddr      string `yaml:"rpc_addr" json:"rpc_addr"`
	allNode      map[string]*Node
	MasterName   string
	Vote         *Vote
	Data         map[uint64][]byte
	Buff         map[uint64]*BuffData
}

type BuffData struct {
	data []byte
	vote int
}

func (n *Node) Change() {
	//switch n.Status {
	//case Follower:
	//	n.Follower
	//case Leader:
	//	n.Leader
	//case Candidate:
	//	n.Candidate
	//}

}

func (n *Node) Sync() {

}
func init() {
}

// TODO 丢失链接的node 需要用什么办法从slice中移除
var mux sync.Mutex

// NewNode 建立一个节点
// 同时要提供其他节点的通讯地址
func NewNode(selfNode *Node) {
	selfNode.allNode = make(map[string]*Node)
	fs, err := os.Open(selfNode.Id + ".yml")
	data, err := io.ReadAll(fs)
	if err != nil {
	}

	// 将YAML数据反序列化为结构体
	err = yaml.Unmarshal(data, &selfNode)
	if err != nil {
	}
	selfNode.Data = make(map[uint64][]byte)
	selfNode.Buff = make(map[uint64]*BuffData)
	selfNode.Net = new(network)
	selfNode.Timer = new(timer)
	selfNode.Vote = new(Vote)
	selfNode.Net.self = selfNode
	selfNode.Timer.self = selfNode
	selfNode.Vote.self = selfNode
	selfNode.Message = make(chan []byte, 1024)
	go selfNode.Timer.Run()
	defer fs.Close()

	// 进入正式的流程
	selfNode.Status = Follower
	selfNode.Channel = make(chan int, 1)
	go func() {
		for {
			select {
			case <-selfNode.Timer.Ticker:
				if selfNode.Status == Follower {
					// 检测超时进入选举流程
					if time.Now().UnixMilli()-selfNode.LastLoseTime > int64(selfNode.Timeout) {
						selfNode.Channel <- Candidate
					}
				} else if selfNode.Status == Leader {
					//
					selfNode.Net.HeartRequest()
					//if len(selfNode.allNode) == 0 { // 重新记录候选者
					//	selfNode.Status = Candidate
					//	selfNode.Channel <- Candidate
					//}
				}
			case n := <-selfNode.Channel:
				if n == Candidate {
					selfNode.Mux.Lock()
					// 暂停信号量 直到执行完成
					selfNode.Timer.Pause()
					// 补充候选者状态 应该暂停其他网络请求
					selfNode.Status = Candidate
					selfNode.TermIndex += 1
					var v int
					for _, node := range selfNode.allNode {
						ne := node
						v += selfNode.Vote.RequestVote(ne)
					}
					// 向其他几点发送投票请求
					// 统计获得票
					if v >= len(selfNode.allNode)/2+1 {
						// 进入领导者状态
						// 通知其他节点建立连接
						selfNode.Status = Leader
						selfNode.Channel <- Leader
						// 向子节点提示建立连接
						for _, node := range selfNode.allNode {
							selfNode.Net.MasterConnect(node.TcpAddr)
						}
					} else {
						// 没有获取到票，退回到follower状态
						// TODO 还有权重的影响
						selfNode.Status = Follower
						selfNode.Channel <- Follower
					}
					selfNode.Mux.Unlock()
					selfNode.Timer.Restart()
					// 如果获取到了 2n+1的票则成为leader
				} else if n == Leader {
					// 向其他人发送心跳并接受到心跳 to :112
					// 校验与其他节点是否建立了连接
					selfNode.IsVote = false
				} else {
					// 最后是跟随者
					selfNode.IsVote = false

					//selfNode.IsVote = false
					// 检查与master节点是不是有通信

				}

			}
		}
	}()
}
