package simple_raft

import (
	"fmt"
	"testing"
)

func TestNewNode(t *testing.T) {
	NewNode("111", []*Node{
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
	fmt.Println(16 & 4)
}
