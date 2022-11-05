// go:doc
// raft算法包括
// 选举 复制 成员变动 日志压缩
// 多个关键词
//  1) 复制状态机 (Replicated state machines) , 只有收到确定指定才会在指定内存地址修改为确定的值,exp: x<-3 => x<-5 ,只有领导接受客户端指令
//  2) 任期 (term),可以看作一个领导人持续的时间段，为一个任期,并且一个无符号自增整型id,称为任期号
//  3) 日志提交 (commit),leader在收到客户端指令时，不会立即执行，而是会产生一个 日志项(log entry),其中包含指令，当log entry被复制到超过一个的节点时,日志项会被leader commit，并执行该项
// 3种角色
//  1) leader 主要是复制日志到其他节点,维持心跳(心跳时间要小于全部节点的超市时间)
//  2) candidate
//  3) follower
// 领导人选举(Leader Election)
//  1) 当没有leader时,第一个超时的follower成为candidate

package simple_raft
