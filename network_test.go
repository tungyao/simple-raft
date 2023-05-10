package simple_raft

import (
	"fmt"
	"log"
	"net"
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
				write, err := conn.Write(pack([]byte("hello world")))
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
}
