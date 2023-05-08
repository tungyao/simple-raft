package simple_raft

import (
	"log"
	"net"
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
	log.Println("发送编号", v.self.TermIndex, v.self.LogIndex)
	send = uint642uint8(v.self.TermIndex, send)
	// 写入log编号
	send = uint642uint8(v.self.LogIndex, send)
	// 写入id
	dial, err := net.Dial("tcp", node.TcpAddr)
	if err != nil {
		log.Fatalln("发送失败", err)
	}
	data := make([]byte, 4)
	_, err = dial.Write(send)

	dial.Read(data)
	// 等待返回
	dial.Close()
	log.Println("获得了多少票", data[3])
	return int(data[3])
}
