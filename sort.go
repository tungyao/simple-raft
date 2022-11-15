package main

type heap interface {
	Add(int) int
	Pop() int
}

type heapMax struct {
	heapSize int
	realSize int
	node     []int
}

func NewHeap(size int) heap {
	var hp heap = &heapMax{node: make([]int, size)}
	return hp
}

// 添加一个新元素
// 最大堆
func (h *heapMax) Add(key int) int {
	h.realSize += 1
	if h.realSize > h.heapSize {
		return -1
	}
	h.node[h.heapSize] = key
	index := h.realSize
	parent := index / 2
	for h.node[index] > h.node[parent] && index > 1 {
		h.node[index], h.node[parent] = h.node[parent], h.node[index]
		index = parent
		parent = parent / 2
	}
	return index
}

// 弹出一个元素
func (h *heapMax) Pop() int {
	if h.realSize < 1 {
		return -1
	}
	removeEl := h.node[1]
	// 将队尾的元素提到第一位
	h.node[1] = h.node[h.realSize]
	h.heapSize -= 1
	index := 1

	for index <= h.realSize/2 {
		left := index * 2
		right := (index * 2) + 1

		if h.node[index] < h.node[left] || h.node[index] < h.node[right] {
			if h.node[left] > h.node[right] {
				h.node[index], h.node[left] = h.node[left], h.node[index]
				index = left
			} else {
				h.node[index], h.node[right] = h.node[right], h.node[index]
				index = right
			}
		}else{
			break
		}

	}

	return 0
}
