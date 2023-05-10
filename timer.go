package simple_raft

// 负责定时相关的东西

import (
	"sync"
	"time"
)

type timer struct {
	Ticker                chan struct{}
	self                  *Node
	ticker                *time.Ticker
	minHeartTimeoutTicker *time.Ticker
	mux                   sync.RWMutex
	same                  bool // 信号量
}

// Run 启动定时器和同步线程状态
// 如果进入选举状态 则不发送定时信号 由same这个控制
func (t *timer) Run() {
	t.Ticker = make(chan struct{})
	t.same = true
	t.ticker = time.NewTicker(time.Millisecond * time.Duration(t.self.Timeout))
	t.minHeartTimeoutTicker = time.NewTicker(time.Hour)
	t.minHeartTimeoutTicker.Stop()
	for {
		select {
		case <-t.ticker.C:
		case <-t.minHeartTimeoutTicker.C:
		}
		t.mux.RLock()
		// equal true , it's not stopping
		if t.same == true {
			t.Ticker <- struct{}{}
		}
		t.mux.RUnlock()

	}
}
func (t *timer) Pause() {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.same = false
}

func (t *timer) Restart() {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.same = true
}

// Start 下面两个方法是系统定时器的 一般不调用
func (t *timer) Start() {
	t.ticker.Reset(time.Millisecond * time.Duration(t.self.Timeout))
	// 选出最小的心跳超时时间
	low := 4095
	t.mux.Lock()
	for _, v := range t.self.allNode {
		if low < v.Timeout {
			low = v.Timeout
		}
	}

	if t.minHeartTimeoutTicker == nil {
		t.minHeartTimeoutTicker = time.NewTicker(time.Millisecond * time.Duration(low) / 2)
	}
	t.minHeartTimeoutTicker.Reset(time.Millisecond * time.Duration(low))
	t.mux.Unlock()
}
func (t *timer) Stop() {
	t.ticker.Stop()
	t.minHeartTimeoutTicker.Stop()
}
