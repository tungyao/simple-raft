package main

// 开始手撸算法
func main() {
	s := NewHeap(5)
	s.Add(1)
	s.Add(3)
	s.Add(2)
	s.Add(4)
	s.Pop()
	println(s)
}
