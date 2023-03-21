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
	Addr         string // 主机地址
	Rate         uint8  // 投票倍率 这个值一般可以忽略
	Id           string // 主机标识
	Status       int
	Net          *network // 网络相关的操作
	Timeout      int      // 心跳间隔
	LastLoseTime int64    // 上次收到心跳时间
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

var pipe chan *Node

// NewNode 建立一个节点
// 同时要提供其他节点的通讯地址
func NewNode(selfId string, node []*Node) *Raft {
	pipe = make(chan *Node, 1)
	rf := new(Raft)
	var selfNode *Node
	for _, v := range node {
		if v.Id == selfId {
			selfNode = v
			break
		}
	}

	fo := &follower{}

	ca := &candidate{}

	le := &leader{}
	go Hub(le, ca, fo)
	go func() {
		time.Sleep(time.Second * 2)
		pipe <- selfNode

	}()
	time.Sleep(time.Second * 10)
	return rf
}
