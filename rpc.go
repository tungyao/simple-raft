package simple_raft

import (
	"context"
	"fmt"
	"github.com/tungyao/simple-raft/pb"
	"google.golang.org/grpc"
	"log"
	"net"
)

// StartRpc 启动grpc监控
func StartRpc(node *Node) {
	grpcServer := grpc.NewServer()
	service1 := &VoteRpcImp{
		self: node,
	}
	pb.RegisterVoteServer(grpcServer, service1)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 10001))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// VoteRpcImp 实现投票接口
type VoteRpcImp struct {
	pb.UnimplementedVoteServer
	self *Node
}

// VoteRequest 收到投票请求
func (v *VoteRpcImp) VoteRequest(ctx context.Context, data *pb.VoteRequestData) (*pb.VoteReplyData, error) {

	mux.Lock()
	defer mux.Unlock()
	if v.self.IsVote == false && data.TermIndex <= v.self.TermIndex || data.LogIndex <= v.self.LogIndex {
		return &pb.VoteReplyData{Get: 0}, nil
	}
	v.self.IsVote = true

	return &pb.VoteReplyData{Get: 1}, nil
}
