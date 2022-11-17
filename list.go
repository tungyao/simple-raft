package main

type list struct {
	next    *list
	key     int
	expired bool
}

func (l *list) Add(key int) {
	ll := l
	for ll.next != nil {
		if ll.next.expired == true {
			ll.next.key = key
			ll.next.expired = false
			return
		}
		ll = ll.next
	}
	ll.next = &list{key: key}
}

func (l *list) Del(key int) {
	ll := l
	for ll.next != nil {
		if ll.next.key == key {
			ll.next.expired = true
			return
		}
		ll = ll.next
	}
}
