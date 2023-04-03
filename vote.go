package simple_raft

// TODO 需要从 raft:116 移植过来

type Vote struct {
	IsThisTermVoting bool   // 在当前任期类是否投过票了
	ThisTerm         uint64 // 当前任期号,仅作为数据记录
}
