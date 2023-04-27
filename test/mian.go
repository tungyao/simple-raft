package main

import (
	"flag"
	sr "github.com/tungyao/simple-raft"
)

var port string
var id string

func main() {
	flag.StringVar(&port, "p", "", "port")
	flag.StringVar(&id, "id", "", "id")
	flag.Parse()
	node := &sr.Node{
		Addr: "127.0.0.1:" + port,
		Id:   id,
	}
	sr.NewNode(node)
	node.Net.Run()

}
