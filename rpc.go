package simple_raft

import (
	"context"
	"github.com/tungyao/simple-raft/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"time"
)

// StartRpc 启动grpc监控
func StartRpc(node *Node) {
	grpcServer := grpc.NewServer()
	service1 := &VoteRpcImp{
		self: node,
	}
	service2 := &EntryRpcImp{
		self: node,
	}
	pb.RegisterVoteServer(grpcServer, service1)
	pb.RegisterEntryServer(grpcServer, service2)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", node.RpcAddr)
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

	v.self.Mux.Lock()
	defer v.self.Mux.Unlock()
	// TODO 这里要修改 任期和日志计算方式 不太对
	log.Println("----", data.TermIndex, v.self.TermIndex, data.LogIndex, v.self.LogIndex, data.TermIndex < v.self.TermIndex || data.LogIndex < v.self.LogIndex)
	if v.self.IsVote == false {
		v.self.IsVote = true
		if v.self.TermIndex > data.TermIndex {
			// term index 更大，优先级更高
			return &pb.VoteReplyData{Get: 0}, nil

		} else if v.self.TermIndex < data.TermIndex {
			// term index 更小，优先级更低
			return &pb.VoteReplyData{Get: 1}, nil

		} else {
			// term index 相同，比较 log index
			if v.self.LogIndex > data.LogIndex {
				// log index 更大，优先级更高
				return &pb.VoteReplyData{Get: 0}, nil

			} else {
				// log index 更小或相等，优先级更低
				return &pb.VoteReplyData{Get: 1}, nil
			}
		}
	}
	log.Println("发送 000 票")
	return &pb.VoteReplyData{Get: 0}, nil
}
func (v *VoteRpcImp) MasterNotice(ctx context.Context, data *pb.VoteMasterRequest) (*pb.VoteMasterReply, error) {
	v.self.Mux.Lock()
	defer v.self.Mux.Unlock()
	v.self.Net.ConnectMaster(v.self.allNode[data.Id].RpcAddr)

	return &pb.VoteMasterReply{Id: data.Id}, nil
}

// EntryRpcImp 向master节点请求其他节点的资料
type EntryRpcImp struct {
	*pb.UnimplementedEntryServer
	self *Node
}

// ClientInit 完成复杂的资料交流
func (e *EntryRpcImp) ClientInit(ctx context.Context, data *pb.EntryClientInitDataRequest) (*pb.EntryClientInitDataReply, error) {
	// 提取远程没有的
	e.self.Mux.Lock()
	defer e.self.Mux.Unlock()
	out := &pb.EntryClientInitDataReply{
		SelfNode: make([]*pb.SelfNode, 0),
		MasterId: e.self.Id,
	}
	for k, v := range e.self.allNode {
		log.Println("----", v.Id)
		var yes bool
		for _, node := range data.SelfNode {
			if k == node.GetId() {
				yes = true
				break
			}
		}
		if yes == false {
			out.SelfNode = append(out.SelfNode, &pb.SelfNode{
				Id:      v.Id,
				TcpAddr: v.TcpAddr,
				RpcAddr: v.RpcAddr,
			})
		}
	}
	for _, i2 := range data.SelfNode {
		if _, ok := e.self.allNode[i2.GetId()]; !ok && i2.GetId() != e.self.Id {
			log.Println("++++", i2.GetId())
			e.self.allNode[i2.GetId()] = &Node{
				Id:      i2.GetId(),
				TcpAddr: i2.GetTcpAddr(),
				RpcAddr: i2.GetRpcAddr(),
				Net:     new(network),
				Timer:   new(timer),
			}
		}
	}
	log.Println(out)
	return out, nil
}
func FastRpcVoteMasterConnect(selfNode *Node, otherNode *Node) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	log.Println(otherNode.Id, "rpc dial", otherNode)
	conn, err := grpc.DialContext(ctx, otherNode.RpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect: %v\n", err)
		delete(otherNode.allNode, otherNode.Id)
		log.Println("已经移除节点了")
		return
	}
	defer conn.Close()

	clt := pb.NewVoteClient(conn)
	notice, err := clt.MasterNotice(ctx, &pb.VoteMasterRequest{Id: selfNode.Id})
	if err != nil {
		log.Println(err, notice)
		return
	}

}
func FastRpcClientInit(selfNode *Node, RpcAddr string) *pb.EntryClientInitDataReply {
	log.Println("向master节点同步资料", selfNode.allNode)
	selfNode.Mux.Lock()
	defer selfNode.Mux.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	log.Println(selfNode.Id, "rpc dial", RpcAddr)
	conn, err := grpc.DialContext(ctx, RpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect: %v\n", err)
		return nil
	}
	defer conn.Close()

	clt := pb.NewEntryClient(conn)
	out := make([]*pb.SelfNode, 0)
	for _, node := range selfNode.allNode {
		log.Println("----", node.Id)
		out = append(out, &pb.SelfNode{
			Id:      node.Id,
			TcpAddr: node.TcpAddr,
			RpcAddr: node.RpcAddr,
		})
	}
	reply, err := clt.ClientInit(ctx, &pb.EntryClientInitDataRequest{
		SelfNode: out,
		SelfId:   selfNode.Id,
	})
	if err != nil {
		log.Printf("did not connect: %v\n", err)
		return nil
	}
	// 写入到本地node中
	for _, i2 := range reply.SelfNode {
		if _, ok := selfNode.allNode[i2.GetId()]; !ok && i2.GetId() != selfNode.Id {
			log.Println("++++", i2.GetId(), selfNode.Id)
			selfNode.allNode[i2.GetId()] = &Node{
				Id:      i2.GetId(),
				TcpAddr: i2.GetTcpAddr(),
				RpcAddr: i2.GetRpcAddr(),
				Net:     new(network),
				Timer:   new(timer),
			}
		}
	}

	return reply
}
