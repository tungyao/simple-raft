package main

import (
	"fmt"
	"unsafe"
)

// Admin 关于unsafe的高级说明 结构体是普遍存在内存对齐的 内存对齐又影响着 指针操作 具体说明方法看下面
type Admin struct {
	Name int8  // 1
	T    int64 // 8
	Age  int16 // 2
}
type Admin2 struct {
	Name int8  // 1
	Age  int16 // 2
	T    int64 // 8

}

func P1() {
	// 指针的通用操作
	a := 10
	fmt.Println(unsafe.Sizeof(Admin2{}), unsafe.Sizeof(Admin{}))
	p := unsafe.Pointer(&a)
	// 这句话实际上干了一件事 获得a的地址然后进行偏移 如果这里不+0 而是+2 +8 就会访问到另外的内存地址
	ptr := (*int64)(unsafe.Pointer(uintptr(p) + 0))
	*ptr = 20
	fmt.Println(a)

	// 下面是结构体骚操作的解释
	// golang中 结构的字段的顺序 有内存对齐的出现  这同一个结构的实际占用内存是不一样的
	// 对齐的规则 从0开始偏移 有2个边界 整体边界和字段边界 可以这样理解 字段边界是 开始地址必须是自身大小的倍数 整体边界是最后的大小 必须是最高字段大小的倍数
	//      24                       16
	// type Admin struct {          type Admin2 struct {
	//	Name int8  // 1  0-1         Name int8  // 1 0-1
	//	T    int64 // 8  8-15         Age  int16 // 2 2-3
	//	Age  int16 // 2  16-17        T    int64 // 8 8-16
	// }                            }
	admin := &Admin2{}
	ptrs := unsafe.Pointer(admin)

	// 在对admin2修改的时候 +8偏移就是 T的开始地址
	age := (*int64)(unsafe.Pointer(uintptr(ptrs) + 8))
	*age = 66
	fmt.Printf("%+v", admin)
}
