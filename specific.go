package simple_raft

import (
	"io"
	"log"
	"math/rand"
	"net"
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
	self    *Node
	Address string
	Rece    chan []byte
	Send    chan []byte
}

func (ne *network) Run() {
	lis, err := net.Listen("tcp", ne.Address)
	if err != nil {
		return
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go func(connInner net.Conn) {
			go func() {
				for {
					data, err := io.ReadAll(connInner)
					if err != nil {
						log.Println(err)
					}
					ne.Rece <- data
				}
			}()

			for {
				select {
				case data := <-ne.Send:
					connInner.Write(data)

				}
			}

		}(conn)
	}
}
func (ne *network) Req() {

}

// VoteRequest 发送自己的
func (ne *network) VoteRequest(node *Node) int {
	log.Printf("进入选举模式: vote: %d ,term: %d ,logindex: %d \n", node.Vote, node.TermIndex, node.LogIndex)
	return rand.Intn(2)
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
