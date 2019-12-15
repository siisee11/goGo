package kth

import (
	"sort"
	//    "fmt"
)

func Solution(array []int, commands [][]int) []int {
	answer := []int{}

	for n := range commands {
		i, j, k := commands[n][0], commands[n][1], commands[n][2]
		new_slice := array[i-1 : j]
		sort.Slice(new_slice, func(a, b int) bool {
			return new_slice[a] < new_slice[b]
		})
		answer = append(answer, new_slice[k-1])
	}

	return answer
}
