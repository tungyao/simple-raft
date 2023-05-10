package main

import (
	"flag"
	sr "github.com/tungyao/simple-raft"
	"log"
)

var port string
var id string
var slave bool
var master string

func main() {
	flag.StringVar(&port, "p", "3000", "port")
	flag.StringVar(&id, "id", "node1", "id")
	flag.StringVar(&master, "master", "127.0.0.1:3000", "127.0.0.1:3000")
	flag.BoolVar(&slave, "slave", false, "slave")
	flag.Parse()
	node := &sr.Node{
		TcpAddr: "127.0.0.1:" + port,
		Id:      id,
	}
	log.Println(id, port, master, slave)
	sr.NewNode(node)
	if slave {
		node.Net.ConnectMaster(&sr.Node{TcpAddr: master})
	}
	node.Net.Run()

}
