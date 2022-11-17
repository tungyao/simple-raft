package main

// 在一个乱序一维数组中 求出其中值最大的一组数组之和
func containsDuplicate(nums []int) bool {
	mp := make(map[int]int8)
	for _, v := range nums {
		if _, ok := mp[v]; ok {
			mp[v] += 1
		} else {
			mp[v] = 1
		}
	}
	for _, v := range mp {
		if v > 1 {
			return true
		}
	}
	return false
}

// nums = [2,7,11,15], target = 9

func twoSum(nums []int, target int) []int {
	for k, v := range nums {
		for j := k + 1; j < len(nums); j++ {
			if v+nums[j] == target {
				return []int{k, j}
			}
		}
	}
	return nil
}
