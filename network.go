package simple_raft

import (
	"log"
	"net"
	"unsafe"
)

// 换一种思路
// 在本地维护多个角色 当时node移交到哪个角色上时 就使用那些方法

// 网络相关的东西
type network struct {
	self    *Node
	Address string
	Rece    chan []byte
	Send    chan []byte
}

// In fact , there's a node list in program
// so, Reload from the yaml file

// 协议 按照byte=8位
// 1 | 2          | 3           | 4           | 5 | 6 | 7 | 8 |
//  是否接续上一报文   操作码
//                   0 发送心跳

func (ne *network) Run() {
	ne.Rece = make(chan []byte, 512)
	ne.Send = make(chan []byte, 512)
	lis, err := net.Listen("tcp", ne.self.Addr)
	if err != nil {
		log.Println(err)
		return
	}
	go func() {
		for {
			select {
			case r := <-ne.Rece:
				log.Println(r)
			case s := <-ne.Send:
				log.Println(s)
			}
		}
	}()
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go func(connInner net.Conn) {
			go func() {
				for {
					data := <-ne.Send
					connInner.Write(data)
				}
			}()
			for {
				var data = make([]byte, 1024)
				n, _ := connInner.Read(data)
				ne.Rece <- data[:n]

				// 加入集群
				if data[3] == 0x1 {
					allNode = append(allNode, &Node{})
				}

			}

		}(conn)
	}
}

// HeartRequest 向其他节点发送心跳
func (ne *network) HeartRequest(nodes []*Node) {
	for _, v := range nodes {
		if v.Id != ne.self.Id {
			// TODO 心跳发送
			log.Println("向", v.Id, "发送心跳")
		}
	}
}

func uint82Uint64[T int | uint8](r ...T) uint64 {
	return uint64(r[0]&0xff) +
		uint64(r[1]&0xff)<<8 +
		uint64(r[2]&0xff)<<16 +
		uint64(r[3]&0xff)<<24 +
		uint64(r[4]&0xff)<<32 +
		uint64(r[5]&0xff)<<40 +
		uint64(r[6]&0xff)<<48 +
		uint64(r[7]&0xff)<<56
}

func uint642uint8[T int | uint8 | uint64](n T, arr unsafe.Pointer) {
	var arr1 = (*[]T)(arr)
	*arr1 = append(*arr1,
		T(n&0xff),
		T((n>>8)&0xff),
		T((n>>16)&0xff),
		T((n>>24)&0xff),
		T((n>>32)&0xff),
		T((n>>40)&0xff),
		T((n>>48)&0xff),
		T((n>>56)&0xff),
	)
}
