package simple_raft

import (
	"log"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
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

	NewNode("222", []*Node{
		&Node{
			Addr:    "1",
			Rate:    0,
			Id:      "111",
			Timeout: 100,
		},
		&Node{
			Addr:    "2",
			Rate:    0,
			Id:      "222",
			Timeout: 200,
		},
		&Node{
			Addr:    "3",
			Rate:    0,
			Id:      "333",
			Timeout: 50,
		},
	})
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

// test pgp signed
