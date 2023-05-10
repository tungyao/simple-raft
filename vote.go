package simple_raft

import (
	"net"
	"time"
)

// TODO 需要从 raft:116 移植过来

type Vote struct {
	isThisTermVoting bool     // 在当前任期类是否投过票了
	thisTerm         uint64   // 当前任期号,仅作为数据记录
	channel          chan int // 等待返回票
	self             *Node
}

// RequestVote 向其他节点请求票
func (v *Vote) RequestVote(node *Node) int {
	send := []byte{0, 0, 6}
	// 写入任期编号
	send = uint642uint8(v.self.TermIndex, send)
	// 写入log编号
	send = uint642uint8(v.self.LogIndex, send)
	// 写入id
	dial, err := net.Dial("tcp", node.TcpAddr)
	if err != nil {
	}
	data := make([]byte, 4)
	_, err = dial.Write(pack(send))
	dial.SetReadDeadline(time.Now().Add(time.Second * 5))
	dial.Read(data)
	// 等待返回
	dial.Close()
	return int(data[3])
}
