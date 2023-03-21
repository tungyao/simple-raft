package simple_raft

import "log"

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

type network struct {
}

func (net *network) Run() {

}
func (net *network) Req() {

}
