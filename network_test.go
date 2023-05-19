package simple_raft

import (
	"fmt"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

func send(conn net.Conn, data []byte) error {
	buffer := pack(data)
	_, err := conn.Write(buffer)
	if err != nil {
		return err
	}
	return nil
}

func TestPack(t *testing.T) {
	ln, err := net.Listen("tcp", ":8888")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err.Error())
			return
		}
		go func(conn net.Conn) {
			err := send(conn, []byte("hello world"))
			if err != nil {
				fmt.Println("Error sending:", err.Error())
				return
			}
			for {
				data, err := receive(conn)
				if err != nil {
					//fmt.Println("Error receiving:", err.Error())
				} else {
					fmt.Println(string(data))
				}
			}
		}(conn)
	}
}

func TestUnpack(t *testing.T) {
	var group sync.WaitGroup
	group.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			conn, err := net.Dial("tcp", "127.0.0.1:8888")
			if err != nil {
				fmt.Println("Error listening:", err.Error())
				return
			}
			defer conn.Close()

			for {
				data, err := receive(conn)
				if err != nil {
					fmt.Println("Error receiving:", err.Error())
					return
				}
				fmt.Println(string(data))
				go func() {
					var i = 0
					for {
						time.Sleep(time.Second)
						i++
						log.Println(1)
						write, err := conn.Write(pack([]byte("abcdefghijklmnopqrstuvwxyz")))
						if err != nil {
							log.Println(write, err)
							return
						}

					}
				}()
				if err != nil {
					fmt.Println("Error sending:", err.Error())
					return
				}
			}
		}()
	}
	group.Wait()
}

func Test_network_Broadcast(t *testing.T) {
	//n := network{}
	conn, err := net.Dial("tcp", "192.168.7.78:3002")
	if err != nil {
		log.Println(err)
	}
	_, err = conn.Write(pack([]byte{0, 0, 9, 104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100}))
	if err != nil {
		log.Println(err)
	}
	conn.Close()
	//n.Broadcast(&ss)
}
func Test_Word(t *testing.T) {
	log.Println([]byte("hello world"))
}
