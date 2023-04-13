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
	"context"
	"github.com/tungyao/simple-raft/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v2"
	"io"
	"log"
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
	Addr         string `yaml:"addr"` // 主机地址
	Rate         uint8  // 投票倍率 这个值一般可以忽略
	Id           string `yaml:"id"` // 主机标识
	Status       int
	Net          *network // 网络相关的操作
	Timeout      int      // 心跳间隔
	LastLoseTime int64    // 上次收到心跳时间
	Channel      chan int
	Message      chan string
	LogIndex     uint64
	TermIndex    uint64
	IsVote       bool // 在当前任期是不是已经投了票
	Timer        *timer
	Mux          sync.RWMutex
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
func init() {
	log.SetFlags(log.Llongfile)
}

// TODO 丢失链接的node 需要用什么办法从slice中移除
var allNode []*Node
var mux sync.Mutex

// NewNode 建立一个节点
// 同时要提供其他节点的通讯地址
func NewNode(selfNode *Node) {
	allNode = make([]*Node, 0)
	fs, err := os.Open("conf.yml")
	data, err := io.ReadAll(fs)

	selfNode.Net.self = selfNode
	selfNode.Timer.self = selfNode
	if err != nil {
		log.Fatalf("无法读取YAML文件：%v", err)
	}
	defer fs.Close()

	// 将YAML数据反序列化为结构体
	err = yaml.Unmarshal(data, &allNode)
	if err != nil {
		log.Fatalf("无法反序列化YAML数据：%v", err)
	}
	log.Println(allNode)
	// 进入正式的流程
	selfNode.Status = Follower
	selfNode.Channel = make(chan int, 1)
	selfNode.Net = &network{self: selfNode}
	go StartRpc(selfNode)
	go func() {
		for {
			select {
			case <-selfNode.Timer.Ticker:
				if selfNode.Status == Follower {
					// 进入选举流程
					selfNode.Channel <- Candidate
				} else if selfNode.Status == Leader {
					//
					selfNode.Net.HeartRequest(allNode)

				}
			case n := <-selfNode.Channel:
				log.Println("接受到channel", n)
				if n == Candidate {
					mux.Lock()
					// 暂停信号量 直到执行完成
					selfNode.Timer.Pause()

					log.Println("进入候选者状态")
					// 补充候选者状态 应该暂停其他网络请求
					selfNode.Status = Candidate
					selfNode.TermIndex += 1
					var group sync.WaitGroup
					var replyData = make([]*pb.VoteReplyData, 0, len(allNode))
					group.Add(len(allNode))
					for _, node := range allNode {
						n := node
						go func() {
							defer group.Done()
							// 超时处理
							ctx, cancel := context.WithTimeout(context.Background(), time.Second)
							defer cancel()
							conn, err := grpc.DialContext(ctx, n.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
							if err != nil {
								log.Printf("did not connect: %v\n", err)
								return
							}
							defer conn.Close()

							clt := pb.NewVoteClient(conn)
							reply, err := clt.VoteRequest(ctx, &pb.VoteRequestData{
								TermIndex: selfNode.TermIndex,
								LogIndex:  selfNode.LogIndex,
							})
							replyData = append(replyData, reply)
						}()

					}
					group.Wait()
					// 向其他几点发送投票请求
					var v uint32
					// 统计获得票
					for _, datum := range replyData {
						if datum != nil {
							v += datum.Get
						}
					}
					log.Println("统计票", v)
					if v >= uint32(len(allNode)/2+1) {
						log.Println(selfNode.Id, "总共获取", v, "超过", uint32(len(allNode)/2+1))
						// 进入领导者状态
						selfNode.Status = Leader
						mux.Unlock()
						selfNode.Channel <- Leader

					} else {
						mux.Unlock()
						// 没有获取到票，退回到follower状态
						// TODO 还有权重的影响
						selfNode.Status = Follower
						selfNode.Channel <- Follower
					}
					selfNode.Timer.Restart()
					// 如果获取到了 2n+1的票则成为leader
				} else if n == Leader {
					// 向其他人发送心跳并接受到心跳 to :112
				} else {
					// 最后是跟随者
				}

			}
		}
	}()
}

type timer struct {
	Ticker                chan struct{}
	self                  *Node
	ticker                *time.Ticker
	minHeartTimeoutTicker *time.Ticker
	mux                   sync.RWMutex
	same                  bool // 信号量
}

// Run 启动定时器和同步线程状态
// 如果进入选举状态 则不发送定时信号 由same这个控制
func (t *timer) Run() {
	t.Ticker = make(chan struct{})
	t.same = true
	t.ticker = time.NewTicker(time.Second)
	t.minHeartTimeoutTicker = time.NewTicker(time.Second)
	t.minHeartTimeoutTicker.Stop()
	for {
		select {
		case <-t.ticker.C:
		case <-t.minHeartTimeoutTicker.C:

		}
		t.mux.RLock()
		// equal true , it's not stopping
		if t.same == true {
			log.Println("发送ticker")
			t.Ticker <- struct{}{}
		}
		t.mux.RUnlock()

	}
}
func (t *timer) Pause() {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.same = false
}

func (t *timer) Restart() {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.same = true
}

// Start 下面两个方法是系统定时器的 一般不调用
func (t *timer) Start() {
	t.ticker.Reset(time.Second)
	// 选出最小的心跳超时时间
	low := 4095
	mux.Lock()
	for _, v := range allNode {
		if low < v.Timeout {
			low = v.Timeout
		}
	}
	mux.Unlock()

	if t.minHeartTimeoutTicker == nil {
		t.minHeartTimeoutTicker = time.NewTicker(time.Millisecond * time.Duration(low))
	}
	t.minHeartTimeoutTicker.Reset(time.Millisecond * time.Duration(low))

}
func (t *timer) Stop() {
	t.ticker.Stop()
	t.minHeartTimeoutTicker.Stop()
}
