package simple_raft

import (
	"log"
	"net"
	"time"
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
	WaitVote     chan int
	WaitVoteCode uint8
}

// In fact , there's a node list in program
// so, Reload from the yaml file

// 协议 按照byte=8位
// 1 | 2          | 3           | 4           |
//  是否接续上一报文   操作码
//                   0 发送心跳
//                   1 请求票       应该回应编号
//                   2 回应票

func (ne *network) Run() {
	ne.Rece = make(chan []byte, 512)
	ne.Send = make(chan []byte, 512)
	ne.WaitVote = make(chan int)
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
				// 发回等待的
				if r[3] == 0x2 && r[4] == ne.WaitVoteCode {
					ne.WaitVote <- 1
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
			}

		}(conn)
	}
}

// VoteRequest 发送自己的
func (ne *network) VoteRequest(node *Node) int {

	ne.Send <- []byte{0, 0, 1, 128}
	ne.WaitVoteCode = 128
	// 进入等待
	select {
	case <-time.After(time.Second * 10):

	case v := <-ne.WaitVote:
		return v
	}
	return 0
}

// HeartRequest 向其他节点发送心跳
func (ne *network) HeartRequest(nodes []*Node) {
	for _, v := range nodes {
		if v.Id != ne.self.Id {
			log.Println("向", v.Id, "发送心跳")
		}
	}
}

func (ne *network) VoteResponse() {
	// 在一轮任期 只能投一次票
	if ne.self.IsVote == false {
		ne.Send <- []byte("yes")
	}
}
