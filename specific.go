package simple_raft

import (
	"log"
	"net"
	"sync"
	"time"
	"unsafe"
)

// 换一种思路
// 在本地维护多个角色 当时node移交到哪个角色上时 就使用那些方法

func Hub(leader2 *leader, candidate2 *candidate, follower2 *follower) {

}

type leader struct {
}

func (le *leader) Run() {
	select {
	case <-pipe:
		log.Println(1)
	}
}

type candidate struct {
}

func (ca *candidate) Run() {

}

type follower struct {
	Vote int   // 自己有多少票
	Node *Node // Node节点
}

func (fo *follower) Run() {
	select {
	case n := <-pipe:
		log.Println(2)

		if n.Status&Follower == Follower {

		}
	}
}

// 仅是做测试一用
type network struct {
	self         *Node
	Address      string
	Rece         chan []byte
	Send         chan []byte
	WaitVote     chan *WaitVoteData
	WaitVoteCode uint8
	Net          net.Conn
}

// WaitVoteData 接受到的投票数据
type WaitVoteData struct {
	Number   uint8  // 回复编号
	Term     uint64 // 发来的任期编号
	LogIndex uint64 // 日志的索引
}

// In fact , there's a node list in program
// so, Reload from the yaml file

// 协议 按照byte=8位
// 1 | 2          | 3           | 4           | 5 | 6 | 7 | 8 |
//  是否接续上一报文   操作码
//                   0 发送心跳
//                   1 请求票       应该回应编号    任期号5-12位       13-20位是日志最大编号
//                   2 回应票

func (ne *network) Run() {
	ne.Rece = make(chan []byte, 512)
	ne.Send = make(chan []byte, 512)
	ne.WaitVote = make(chan *WaitVoteData)
	lis, err := net.Listen("tcp", ne.self.Addr)
	if err != nil {
		log.Println(err)
		return
	}
	go func() {
		for {
			select {
			case r := <-ne.Rece:
				log.Println(r)


				// 进入投票模式
				if r[3] == 0x2 && r[4] == ne.WaitVoteCode {
					term := uint82Uint64(r[5], r[6], r[7], r[8], r[9], r[10], r[11], r[12])
					logIndex := uint82Uint64(r[13], r[14], r[15], r[16], r[17], r[18], r[19], r[20])
					ne.WaitVote <- &WaitVoteData{Number: r[4], Term: term, LogIndex: logIndex}
				}
			case s := <-ne.Send:
				log.Println(s)
			}
		}
	}()
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go func(connInner net.Conn) {
			go func() {
				for {
					data := <-ne.Send
					connInner.Write(data)
				}
			}()
			for {
				var data = make([]byte, 1024)
				n, _ := connInner.Read(data)
				ne.Rece <- data[:n]

				// 加入集群
				if data[3] == 0x1 {
					allNode = append(allNode, &Node{})
				}

			}

		}(conn)
	}
}

// VoteRequest 发起投票
func (ne *network) VoteRequest(node *Node) int {

	ne.Send <- []byte{0, 0, 1, 128}
	ne.WaitVoteCode = 128
	// 向其他node发送选票请求
	var group sync.WaitGroup
	group.Add(len(allNode))
	for _, v := range allNode {
		nd := v
		go func() {
			nd.Net.Send <-
		}()
		// TODO 这里有个前置条件 需要保存每一个node的连接标识
	}
	return 0
}

// HeartRequest 向其他节点发送心跳
func (ne *network) HeartRequest(nodes []*Node) {
	for _, v := range nodes {
		if v.Id != ne.self.Id {
			// TODO 心跳发送
			log.Println("向", v.Id, "发送心跳")
		}
	}
}

func (ne *network) VoteResponse() int {
	//TODO 在这里处理投票的逻辑 主要是对比任期和日志最大的编号 而且在一个任期内只能投一次
	select {
	case <-time.After(time.Second * 10):

	case v := <-ne.WaitVote:
		log.Println(v)
		if ne.self.IsVote == false && v.LogIndex > ne.self.LogIndex && v.Term > ne.self.TermIndex {
			return 1
		}
		return 0
	}
	return 0

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
