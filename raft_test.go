package simple_raft

import (
	"context"
	"log"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

func TestNewNode(t *testing.T) {
	//
	//go func() {
	//	NewNode("111", []*Node{
	//		&Node{
	//			Addr:    "1",
	//			Rate:    0,
	//			Id:      "111",
	//			Timeout: 100,
	//		},
	//		&Node{
	//			Addr:    "2",
	//			Rate:    0,
	//			Id:      "222",
	//			Timeout: 200,
	//		},
	//		&Node{
	//			Addr:    "3",
	//			Rate:    0,
	//			Id:      "333",
	//			Timeout: 50,
	//		},
	//	})
	//}()
	//
	//go func() {
	//	NewNode("222", []*Node{
	//		&Node{
	//			Addr:    "1",
	//			Rate:    0,
	//			Id:      "111",
	//			Timeout: 100,
	//		},
	//		&Node{
	//			Addr:    "2",
	//			Rate:    0,
	//			Id:      "222",
	//			Timeout: 200,
	//		},
	//		&Node{
	//			Addr:    "3",
	//			Rate:    0,
	//			Id:      "333",
	//			Timeout: 50,
	//		},
	//	})
	//}()
}
func TestItoa(t *testing.T) {
	var a int32 = 1
	go func() {
		atomic.AddInt32(&a, 1)
		log.Println(a)
	}()
	go func() {
		atomic.AddInt32(&a, 1)
		log.Println(a)
	}()
	atomic.AddInt32(&a, 1)
	log.Println(a)
	time.Sleep(time.Hour)
}
func TestSelect(t *testing.T) {
	ch := make(chan int)
	for {
		select {
		case <-time.After(time.Second * 3):
			log.Println("second")
			go func() {
				ch <- rand.Intn(3)
			}()
		case n := <-ch:
			switch n {
			case 1:
				log.Println("1", n)
				go func() {

				}()
			case 2:
				log.Println("2", n)
			}
		}
	}
}

func TestNode1(t *testing.T) {
	node := &Node{
		Addr:         "127.0.0.1:3000",
		Rate:         0,
		Id:           "1",
		Status:       0,
		Net:          new(network),
		Timeout:      100,
		LastLoseTime: 0,
		Channel:      nil,
		Message:      nil,
		LogIndex:     0,
		TermIndex:    0,
		IsVote:       false,
		Timer:        new(timer),
	}
	NewNode(node)
	node.Timer.Run()
	node.Net.Run()
}
func TestNode2(t *testing.T) {

}
func TestNode3(t *testing.T) {

}

func TestPlus(t *testing.T) {
	var arr1 = make([]int, 0)

	a := 21565
	//arr1 = append(arr1, a&0xff)
	//arr1 = append(arr1, (a>>8)&0xff)
	//arr1 = append(arr1, (a>>16)&0xff)
	//arr1 = append(arr1, (a>>24)&0xff)
	uint642uint8(a, unsafe.Pointer(&arr1))
	log.Println(arr1)
	b :=
		uint82Uint64(arr1[0], arr1[1], arr1[2], arr1[3], arr1[4], arr1[5], arr1[6], arr1[7])
	log.Println(b)

}
func TestTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	go func(ctx2 context.Context) {
	}(ctx)
	select {
	case <-ctx.Done():
		cancel()

	}
}
