package main

import "fmt"

// 动态规划第一题是特别的经典
// 在一个乱序一维数组中 求出其中值最大的一组数组之和
// nums = [-2,1,-3,4,-1,2,1,-5,4]

func maxSubArray(nums []int) int {
	max := nums[0]
	for i := 1; i < len(nums); i++ {
		if nums[i]+nums[i-1] > nums[i] {
			nums[i] += nums[i-1]
		}
		if nums[i] > max {
			max = nums[i]
		}
		fmt.Println(max)
	}
	return max
}
