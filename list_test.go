package main

import "testing"

func Test_list_Add(t *testing.T) {
	l := new(list)
	l.key = -1
	l.Add(1)
	l.Add(2)
	l.Add(3)
	l.Del(2)
	l.Add(4)
}
